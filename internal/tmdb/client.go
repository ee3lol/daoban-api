package tmdb

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"ee3lol/daoban-api/internal/config"
	"ee3lol/daoban-api/internal/scraper"
)

func FetchDetails(cfg *config.Config, mediaType, tmdbId string) (*scraper.MediaDetails, error) {
	url := fmt.Sprintf("%s/%s/%s", TMDB_BASE_URL, mediaType, tmdbId)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+cfg.TMDBToken)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("TMDB returned status %d", resp.StatusCode)
	}

	var data struct {
		Title        string `json:"title"`
		Name         string `json:"name"`
		Overview     string `json:"overview"`
		PosterPath   string `json:"poster_path"`
		BackdropPath string `json:"backdrop_path"`
		ReleaseDate  string `json:"release_date"`
		FirstAirDate string `json:"first_air_date"`
		Status       string `json:"status"`
		VoteAverage  float64 `json:"vote_average"`
		Genres       []struct{ Name string `json:"name"` } `json:"genres"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	title := data.Title
	if title == "" {
		title = data.Name
	}
	date := data.ReleaseDate
	if date == "" {
		date = data.FirstAirDate
	}
	year := 0
	if len(date) >= 4 {
		year, _ = strconv.Atoi(date[:4])
	}
	poster := ""
	if data.PosterPath != "" {
		poster = "https://image.tmdb.org/t/p/w500" + data.PosterPath
	}
	cover := ""
	if data.BackdropPath != "" {
		cover = "https://image.tmdb.org/t/p/w1280" + data.BackdropPath
	}
	var genres []string
	for _, g := range data.Genres {
		genres = append(genres, g.Name)
	}

	status := "Unknown"
	switch strings.ToLower(data.Status) {
	case "returning series":
		status = "Ongoing"
	case "ended", "canceled", "released":
		status = "Completed"
	}

	return &scraper.MediaDetails{
		ID:          tmdbId,
		Title:       title,
		Description: data.Overview,
		Poster:      poster,
		Cover:       cover,
		Type:        mediaType,
		ReleaseYear: year,
		Status:      status,
		Genres:      genres,
		Rating:      data.VoteAverage,
		Episodes:    []scraper.Episode{},
		Sources:     []scraper.StreamSource{},
	}, nil
}
