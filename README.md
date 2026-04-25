# Groupie Tracker Geolocalization

Groupie Tracker Geolocalization is a Go web application that fetches artist, location, date, and relation data from the public Groupie Tracker API, converts concert addresses into geographic coordinates, and displays each artist's tour stops on an interactive map.

The artist detail page shows geocoded markers and connects them with a line in historical order using the earliest concert date for each location.

## Features

- Backend written in Go using only standard Go packages
- Artist search with exact matches and similar suggestions
- Artist detail pages with lineup, API locations, concert dates, and relation timeline
- Server-side geocoding through the Nominatim/OpenStreetMap API
- Local fallback coordinates for the school audit dataset so maps remain stable if geocoding is unavailable
- Interactive Leaflet/OpenStreetMap frontend map with markers and connected route lines
- Styled error pages and HTTP method/status handling
- Unit tests for API handling, geocoding, route ordering, templates, and search helpers

## Tech Stack

- Go
- HTML templates
- CSS
- Vanilla JavaScript
- Groupie Tracker public API
- Nominatim/OpenStreetMap geocoding API
- Leaflet map display

## Project Structure

```text
.
|-- api/
|   |-- api.go
|   |-- api_test.go
|   |-- geocode.go
|   `-- geocode_test.go
|-- handlers/
|   |-- handlers.go
|   `-- handlers_test.go
|-- models/
|   `-- models.go
|-- static/
|   |-- map.js
|   `-- style.css
|-- templates/
|   |-- artist.html
|   |-- error.html
|   `-- index.html
|-- go.mod
|-- main.go
`-- README.md
```

## Run

```bash
go run .
```

Open:

```text
http://localhost:8080
```

Search for artists such as `Queen`, `ACDC`, `Imagine Dragons`, `Guns N' Roses`, `Post Malone`, or `Red Hot Chili Peppers`, then open the artist detail page to see mapped concert markers and the connected route.

## Test

```bash
go test ./...
```

PowerShell with a local Go cache:

```powershell
$env:GOCACHE = Join-Path (Get-Location) ".gocache"; go test ./...
```

## Routes

- `/` shows the geolocalization dashboard and artist search
- `/artist?id=<id>` shows one artist with geocoded concert markers
- `/search?q=<name>` searches artists by name
- Unknown routes return a styled `404 Not Found` page

## Notes

The Go backend performs address-to-coordinate conversion with a geocoding API. Leaflet is used only in the browser to render the map from coordinates prepared by the server.

## Authors

- Kairzhan
- Asylbek
