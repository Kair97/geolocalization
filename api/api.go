package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"groupie-tracker/models"
	"net/http"
	"time"
)

var ErrArtistNotFound = errors.New("artist not found")

var BaseURL = "https://groupietrackers.herokuapp.com/api"

var client = &http.Client{
	Timeout: 10 * time.Second,
}

type statusError struct {
	URL    string
	Code   int
	Status string
}

func (e *statusError) Error() string {
	return fmt.Sprintf("bad status from %s: %s", e.URL, e.Status)
}

func fetch(url string, target interface{}) error {
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &statusError{
			URL:    url,
			Code:   resp.StatusCode,
			Status: resp.Status,
		}
	}

	err = json.NewDecoder(resp.Body).Decode(target)
	if err != nil {
		return fmt.Errorf("failed to decode response from %s: %w", url, err)
	}

	return nil

}

func GetArtists() ([]models.Artist, error) {
	var artists []models.Artist

	err := fetch(BaseURL+"/artists", &artists)
	if err != nil {
		return nil, err
	}

	return artists, nil
}

func GetArtist(id int) (models.Artist, error) {
	var artist models.Artist

	err := fetch(fmt.Sprintf("%s/artists/%d", BaseURL, id), &artist)
	if err != nil {
		var statusErr *statusError
		if errors.As(err, &statusErr) && statusErr.Code == http.StatusNotFound {
			return models.Artist{}, fmt.Errorf("%w: %d", ErrArtistNotFound, id)
		}
		return models.Artist{}, err
	}

	if artist.ID != id || artist.Name == "" {
		return models.Artist{}, fmt.Errorf("%w: %d", ErrArtistNotFound, id)
	}

	return artist, nil
}

func GetLocations() (models.LocationIndex, error) {
	var locations models.LocationIndex

	err := fetch(BaseURL+"/locations", &locations)
	if err != nil {
		return models.LocationIndex{}, err
	}

	return locations, nil

}

func GetDates() (models.DateIndex, error) {
	var dates models.DateIndex

	err := fetch(BaseURL+"/dates", &dates)
	if err != nil {
		return models.DateIndex{}, err
	}

	return dates, nil
}

func GetRelations() (models.RelationIndex, error) {
	var relations models.RelationIndex

	err := fetch(BaseURL+"/relation", &relations)
	if err != nil {
		return models.RelationIndex{}, err
	}

	return relations, nil
}
