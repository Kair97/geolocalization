package handlers

import (
	"errors"
	"fmt"
	"groupie-tracker/api"
	"groupie-tracker/models"
	"html/template"
	"net/http"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
)

type IndexPageData struct {
	Query   string
	Artists []ArtistCard
	Message string
	Visuals VisualizationData
}

type ArtistCard struct {
	models.Artist
	LocationCount int
	DateCount     int
}

type ConcertStop struct {
	Location  string
	Dates     []string
	DateCount int
	BarWidth  int
}

type ArtistPageData struct {
	Artist    models.Artist
	Locations []string
	Dates     []string
	Concerts  []ConcertStop
	Summary   ArtistSummary
}

type VisualizationData struct {
	TotalArtists      int
	TotalLocations    int
	TotalConcertDates int
	AverageMembers    string
	EarliestCreation  int
	LatestCreation    int
	Decades           []DecadeBucket
	TopTouringArtists []ArtistMetric
}

type DecadeBucket struct {
	Label    string
	Count    int
	BarWidth int
}

type ArtistMetric struct {
	ID            int
	Name          string
	LocationCount int
	DateCount     int
	BarWidth      int
}

type ArtistSummary struct {
	LocationCount int
	DateCount     int
	ConcertCount  int
	PeakStop      ConcertStop
}

type ErrorPageData struct {
	StatusCode int
	Title      string
	Message    string
	ActionURL  string
	ActionText string
}

var templateFuncs = template.FuncMap{
	"formatDate":     formatDate,
	"formatLocation": formatLocation,
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if r.URL.Path != "/" {
		renderError(w, http.StatusNotFound, "Page not found", "The page you requested does not exist. Return to the visual dashboard to keep exploring artists.", "/", "Back to dashboard")
		return
	}

	artists, locations, dates, err := loadIndexResources()
	if err != nil {
		http.Error(w, "Failed to fetch artists", http.StatusInternalServerError)
		return
	}

	err = renderIndex(w, IndexPageData{
		Artists: buildArtistCards(artists, locations.Index, dates.Index),
		Visuals: buildVisualizationData(artists, locations.Index, dates.Index),
	})
	if err != nil {
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
	}
}

func ArtistHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "Missing artist ID", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid artist ID", http.StatusBadRequest)
		return
	}

	artist, err := api.GetArtist(id)
	if err != nil {
		if errors.Is(err, api.ErrArtistNotFound) {
			http.Error(w, "Artist not found", http.StatusNotFound)
			return
		}

		http.Error(w, "Failed to fetch artist", http.StatusInternalServerError)
		return
	}

	locations, dates, relations, err := loadArtistResources()
	if err != nil {
		http.Error(w, "Failed to fetch artist details", http.StatusInternalServerError)
		return
	}

	locationMap := buildLocationMap(locations.Index)
	dateMap := buildDateMap(dates.Index)

	data := ArtistPageData{
		Artist:    artist,
		Locations: locationMap[id],
		Dates:     dateMap[id],
		Concerts:  buildConcerts(id, relations.Index),
	}
	data.Summary = buildArtistSummary(data.Locations, data.Dates, data.Concerts)

	err = renderTemplate(w, "templates/artist.html", data)
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
	}
}

func SearchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := strings.TrimSpace(r.URL.Query().Get("q"))

	artists, locations, dates, err := loadIndexResources()
	if err != nil {
		http.Error(w, "Failed to fetch artists", http.StatusInternalServerError)
		return
	}

	if query == "" {
		err = renderIndex(w, IndexPageData{
			Artists: buildArtistCards(artists, locations.Index, dates.Index),
			Message: "Type an artist name to search the collection.",
			Visuals: buildVisualizationData(artists, locations.Index, dates.Index),
		})
		if err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
		}
		return
	}

	data := IndexPageData{
		Query:   query,
		Visuals: buildVisualizationData(artists, locations.Index, dates.Index),
	}

	matches := findArtists(artists, query)
	data.Artists = buildArtistCards(matches, locations.Index, dates.Index)
	if len(data.Artists) > 0 {
		data.Message = fmt.Sprintf("Showing %d result%s for \"%s\".", len(data.Artists), plural(len(data.Artists)), query)
	} else {
		suggestions := suggestArtists(artists, query)
		data.Artists = buildArtistCards(suggestions, locations.Index, dates.Index)
		if len(data.Artists) > 0 {
			data.Message = fmt.Sprintf("No exact match for \"%s\". Similar artists you may mean:", query)
		} else {
			data.Message = fmt.Sprintf("No artist found for \"%s\".", query)
		}
	}

	err = renderIndex(w, data)
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
	}
}

func renderIndex(w http.ResponseWriter, data IndexPageData) error {
	return renderTemplate(w, "templates/index.html", data)
}

func renderTemplate(w http.ResponseWriter, file string, data interface{}) error {
	tmpl, err := template.New(filepath.Base(file)).Funcs(templateFuncs).ParseFiles(file)
	if err != nil {
		return err
	}

	return tmpl.Execute(w, data)
}

func renderError(w http.ResponseWriter, statusCode int, title, message, actionURL, actionText string) {
	w.WriteHeader(statusCode)
	err := renderTemplate(w, "templates/error.html", ErrorPageData{
		StatusCode: statusCode,
		Title:      title,
		Message:    message,
		ActionURL:  actionURL,
		ActionText: actionText,
	})
	if err != nil {
		http.Error(w, http.StatusText(statusCode), statusCode)
	}
}

func loadIndexResources() ([]models.Artist, models.LocationIndex, models.DateIndex, error) {
	var (
		artists   []models.Artist
		locations models.LocationIndex
		dates     models.DateIndex
	)

	var wg sync.WaitGroup
	errCh := make(chan error, 3)

	wg.Add(3)

	go func() {
		defer wg.Done()
		var err error
		artists, err = api.GetArtists()
		if err != nil {
			errCh <- err
		}
	}()

	go func() {
		defer wg.Done()
		var err error
		locations, err = api.GetLocations()
		if err != nil {
			errCh <- err
		}
	}()

	go func() {
		defer wg.Done()
		var err error
		dates, err = api.GetDates()
		if err != nil {
			errCh <- err
		}
	}()

	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			return nil, models.LocationIndex{}, models.DateIndex{}, err
		}
	}

	return artists, locations, dates, nil
}

func loadArtistResources() (models.LocationIndex, models.DateIndex, models.RelationIndex, error) {
	var (
		locations models.LocationIndex
		dates     models.DateIndex
		relations models.RelationIndex
	)

	var wg sync.WaitGroup
	errCh := make(chan error, 3)

	wg.Add(3)

	go func() {
		defer wg.Done()
		var err error
		locations, err = api.GetLocations()
		if err != nil {
			errCh <- err
		}
	}()

	go func() {
		defer wg.Done()
		var err error
		dates, err = api.GetDates()
		if err != nil {
			errCh <- err
		}
	}()

	go func() {
		defer wg.Done()
		var err error
		relations, err = api.GetRelations()
		if err != nil {
			errCh <- err
		}
	}()

	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			return models.LocationIndex{}, models.DateIndex{}, models.RelationIndex{}, err
		}
	}

	return locations, dates, relations, nil
}

func buildArtistCards(artists []models.Artist, locations []models.Location, dates []models.Date) []ArtistCard {
	locationMap := buildLocationMap(locations)
	dateMap := buildDateMap(dates)

	cards := make([]ArtistCard, 0, len(artists))
	for _, artist := range artists {
		cards = append(cards, ArtistCard{
			Artist:        artist,
			LocationCount: len(locationMap[artist.ID]),
			DateCount:     len(dateMap[artist.ID]),
		})
	}

	return cards
}

func buildVisualizationData(artists []models.Artist, locations []models.Location, dates []models.Date) VisualizationData {
	locationMap := buildLocationMap(locations)
	dateMap := buildDateMap(dates)

	data := VisualizationData{
		TotalArtists:     len(artists),
		EarliestCreation: 0,
		LatestCreation:   0,
	}

	totalMembers := 0
	decadeCounts := make(map[int]int)
	for i, artist := range artists {
		totalMembers += len(artist.Members)
		if i == 0 || artist.CreationDate < data.EarliestCreation {
			data.EarliestCreation = artist.CreationDate
		}
		if artist.CreationDate > data.LatestCreation {
			data.LatestCreation = artist.CreationDate
		}
		if artist.CreationDate > 0 {
			decadeCounts[(artist.CreationDate/10)*10]++
		}
	}

	for _, location := range locations {
		data.TotalLocations += len(location.Locations)
	}
	for _, date := range dates {
		data.TotalConcertDates += len(date.Dates)
	}
	if len(artists) > 0 {
		data.AverageMembers = fmt.Sprintf("%.1f", float64(totalMembers)/float64(len(artists)))
	}

	data.Decades = buildDecadeBuckets(decadeCounts)
	data.TopTouringArtists = buildTopTouringArtists(artists, locationMap, dateMap)

	return data
}

func buildDecadeBuckets(decadeCounts map[int]int) []DecadeBucket {
	decades := make([]int, 0, len(decadeCounts))
	maxCount := 0
	for decade, count := range decadeCounts {
		decades = append(decades, decade)
		if count > maxCount {
			maxCount = count
		}
	}
	sort.Ints(decades)

	buckets := make([]DecadeBucket, 0, len(decades))
	for _, decade := range decades {
		count := decadeCounts[decade]
		buckets = append(buckets, DecadeBucket{
			Label:    fmt.Sprintf("%ds", decade),
			Count:    count,
			BarWidth: relativeWidth(count, maxCount),
		})
	}

	return buckets
}

func buildTopTouringArtists(artists []models.Artist, locationMap, dateMap map[int][]string) []ArtistMetric {
	metrics := make([]ArtistMetric, 0, len(artists))
	maxLocations := 0
	for _, artist := range artists {
		locationCount := len(locationMap[artist.ID])
		dateCount := len(dateMap[artist.ID])
		if locationCount > maxLocations {
			maxLocations = locationCount
		}
		metrics = append(metrics, ArtistMetric{
			ID:            artist.ID,
			Name:          artist.Name,
			LocationCount: locationCount,
			DateCount:     dateCount,
		})
	}

	sort.Slice(metrics, func(i, j int) bool {
		if metrics[i].LocationCount == metrics[j].LocationCount {
			if metrics[i].DateCount == metrics[j].DateCount {
				return metrics[i].Name < metrics[j].Name
			}
			return metrics[i].DateCount > metrics[j].DateCount
		}
		return metrics[i].LocationCount > metrics[j].LocationCount
	})

	limit := 5
	if len(metrics) < limit {
		limit = len(metrics)
	}
	metrics = metrics[:limit]
	for i := range metrics {
		metrics[i].BarWidth = relativeWidth(metrics[i].LocationCount, maxLocations)
	}

	return metrics
}

func buildLocationMap(locations []models.Location) map[int][]string {
	locationMap := make(map[int][]string, len(locations))

	for _, location := range locations {
		sortedLocations := append([]string(nil), location.Locations...)
		sort.Slice(sortedLocations, func(i, j int) bool {
			return formatLocation(sortedLocations[i]) < formatLocation(sortedLocations[j])
		})
		locationMap[location.ID] = sortedLocations
	}

	return locationMap
}

func buildDateMap(dates []models.Date) map[int][]string {
	dateMap := make(map[int][]string, len(dates))

	for _, date := range dates {
		sortedDates := append([]string(nil), date.Dates...)
		sortDates(sortedDates)
		dateMap[date.ID] = sortedDates
	}

	return dateMap
}

func buildConcerts(id int, relations []models.Relation) []ConcertStop {
	for _, rel := range relations {
		if rel.ID != id {
			continue
		}

		concerts := make([]ConcertStop, 0, len(rel.DatesLocations))
		for location, dates := range rel.DatesLocations {
			sortedDates := append([]string(nil), dates...)
			sortDates(sortedDates)
			concerts = append(concerts, ConcertStop{
				Location:  location,
				Dates:     sortedDates,
				DateCount: len(sortedDates),
			})
		}

		sort.Slice(concerts, func(i, j int) bool {
			return formatLocation(concerts[i].Location) < formatLocation(concerts[j].Location)
		})

		maxDates := 0
		for _, concert := range concerts {
			if concert.DateCount > maxDates {
				maxDates = concert.DateCount
			}
		}
		for i := range concerts {
			concerts[i].BarWidth = relativeWidth(concerts[i].DateCount, maxDates)
		}

		return concerts
	}

	return nil
}

func buildArtistSummary(locations, dates []string, concerts []ConcertStop) ArtistSummary {
	summary := ArtistSummary{
		LocationCount: len(locations),
		DateCount:     len(dates),
		ConcertCount:  len(concerts),
	}

	for _, concert := range concerts {
		if concert.DateCount > summary.PeakStop.DateCount {
			summary.PeakStop = concert
		}
	}

	return summary
}

func relativeWidth(value, maxValue int) int {
	if value <= 0 || maxValue <= 0 {
		return 0
	}

	width := value * 100 / maxValue
	if width < 8 {
		return 8
	}

	return width
}

func findArtists(artists []models.Artist, query string) []models.Artist {
	normalizedQuery := normalize(query)
	if normalizedQuery == "" {
		return nil
	}

	var matches []models.Artist
	for _, artist := range artists {
		if strings.Contains(normalize(artist.Name), normalizedQuery) {
			matches = append(matches, artist)
		}
	}

	return matches
}

func suggestArtists(artists []models.Artist, query string) []models.Artist {
	normalizedQuery := normalize(query)
	queryTokens := uniqueTokens(query)

	type scoredArtist struct {
		artist models.Artist
		score  int
	}

	var scored []scoredArtist
	for _, artist := range artists {
		name := normalize(artist.Name)
		score := tokenMatchScore(queryTokens, uniqueTokens(artist.Name))

		if distance := levenshtein(normalizedQuery, name); distance <= 2 {
			score += 3
		}

		if score >= suggestionThreshold(len(queryTokens)) {
			scored = append(scored, scoredArtist{
				artist: artist,
				score:  score,
			})
		}
	}

	sort.Slice(scored, func(i, j int) bool {
		if scored[i].score == scored[j].score {
			return scored[i].artist.Name < scored[j].artist.Name
		}
		return scored[i].score > scored[j].score
	})

	limit := 6
	if len(scored) < limit {
		limit = len(scored)
	}

	suggestions := make([]models.Artist, 0, limit)
	for i := 0; i < limit; i++ {
		suggestions = append(suggestions, scored[i].artist)
	}

	return suggestions
}

func tokenMatchScore(queryTokens, nameTokens []string) int {
	score := 0
	for _, queryToken := range queryTokens {
		for _, nameToken := range nameTokens {
			switch {
			case queryToken == nameToken:
				score += 3
				goto nextToken
			case strings.Contains(nameToken, queryToken), strings.Contains(queryToken, nameToken):
				score += 2
				goto nextToken
			}
		}
	nextToken:
	}

	return score
}

func suggestionThreshold(tokenCount int) int {
	switch {
	case tokenCount >= 3:
		return 4
	case tokenCount == 2:
		return 3
	default:
		return 2
	}
}

func uniqueTokens(value string) []string {
	parts := strings.FieldsFunc(strings.ToLower(value), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})

	seen := make(map[string]bool)
	tokens := make([]string, 0, len(parts))
	for _, part := range parts {
		if part == "" || seen[part] {
			continue
		}

		seen[part] = true
		tokens = append(tokens, part)
	}

	return tokens
}

func normalize(value string) string {
	return strings.Join(uniqueTokens(value), " ")
}

func formatLocation(location string) string {
	if location == "" {
		return ""
	}

	segments := strings.Split(location, "-")
	formatted := make([]string, 0, len(segments))
	for _, segment := range segments {
		segment = strings.ReplaceAll(segment, "_", " ")
		words := strings.Fields(segment)
		for i, word := range words {
			words[i] = titleWord(word)
		}
		formatted = append(formatted, strings.Join(words, " "))
	}

	return strings.Join(formatted, ", ")
}

func titleWord(word string) string {
	if len(word) <= 3 {
		return strings.ToUpper(word)
	}

	runes := []rune(strings.ToLower(word))
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

func formatDate(value string) string {
	date, err := time.Parse("02-01-2006", value)
	if err != nil {
		return value
	}

	return date.Format("02 Jan 2006")
}

func sortDates(dates []string) {
	sort.Slice(dates, func(i, j int) bool {
		left, leftErr := time.Parse("02-01-2006", dates[i])
		right, rightErr := time.Parse("02-01-2006", dates[j])
		if leftErr == nil && rightErr == nil {
			return left.Before(right)
		}

		return dates[i] < dates[j]
	})
}

func plural(count int) string {
	if count == 1 {
		return ""
	}

	return "s"
}

func levenshtein(a, b string) int {
	if a == b {
		return 0
	}

	if a == "" {
		return len([]rune(b))
	}

	if b == "" {
		return len([]rune(a))
	}

	aRunes := []rune(a)
	bRunes := []rune(b)

	previous := make([]int, len(bRunes)+1)
	for j := range previous {
		previous[j] = j
	}

	for i, aRune := range aRunes {
		current := make([]int, len(bRunes)+1)
		current[0] = i + 1

		for j, bRune := range bRunes {
			cost := 0
			if aRune != bRune {
				cost = 1
			}

			current[j+1] = min3(
				current[j]+1,
				previous[j+1]+1,
				previous[j]+cost,
			)
		}

		previous = current
	}

	return previous[len(bRunes)]
}

func min3(a, b, c int) int {
	if a < b && a < c {
		return a
	}

	if b < c {
		return b
	}

	return c
}
