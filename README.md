# Groupie Tracker Visualizations

Groupie Tracker Visualizations is a Go web application that fetches public music data from the Groupie Tracker API and presents it as a responsive, accessible dashboard. The project focuses on clear visual hierarchy, consistent interaction patterns, readable contrast, and graceful error handling.

## Features

- Responsive artist dashboard with aggregate API statistics
- Visual charts for artist formation decades and top touring artists
- Artist cards with image, creation year, first album, member count, location count, and concert date count
- Artist detail pages with lineup, locations, dates, concert timeline, and per-stop activity meters
- Server-side search with exact matches and similar-artist suggestions
- Styled 404 page for unknown routes
- Focus states, readable color contrast, consistent spacing, and mobile-friendly layouts
- Tests for API handling, template rendering, search helpers, visualization data, and route errors

## Tech Stack

- Go
- HTML templates
- CSS
- Groupie Tracker public API

## Project Structure

```text
.
|-- api/
|   |-- api.go
|   `-- api_test.go
|-- handlers/
|   |-- handlers.go
|   `-- handlers_test.go
|-- models/
|   `-- models.go
|-- static/
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

## Test

```bash
go test ./...
```

If your environment blocks the default Go build cache, run tests with a local cache:

```bash
GOCACHE=.gocache go test ./...
```

PowerShell:

```powershell
$env:GOCACHE = Join-Path (Get-Location) ".gocache"; go test ./...
```

## Routes

- `/` shows the visual dashboard and all artists
- `/artist?id=<id>` shows one artist with detailed visualized data
- `/search?q=<name>` searches artists by name
- Any unknown route returns a styled `404 Not Found` page

## Design Notes

The interface follows Schneiderman's 8 Golden Rules by using consistent layout patterns, visible feedback, reversible navigation, simple errors, clear route closure, keyboard focus states, reduced memory load through summaries, and user-controlled browsing/searching.

## Authors

- Kairzhan
- Asylbek
