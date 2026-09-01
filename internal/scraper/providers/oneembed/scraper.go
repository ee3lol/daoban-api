package oneembed

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"ee3lol/daoban-api/internal/scraper"
)

const (
	BaseURL    = "https://1embed.cc"
	TokenTTLMs = 60 * 1000 // 60 seconds
)

type Scraper struct {
	client         *http.Client
	token          string
	tokenExpiresAt time.Time
	mu             sync.Mutex
	apiBaseURL     string
}

func NewScraper(apiBaseURL string) *Scraper {
	return &Scraper{
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
		apiBaseURL: apiBaseURL,
	}
}

func (s *Scraper) getToken() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.token != "" && time.Now().Before(s.tokenExpiresAt) {
		return s.token, nil
	}

	req, err := http.NewRequest("GET", BaseURL+"/api/token", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Referer", BaseURL+"/")
	req.Header.Set("Origin", BaseURL)
	req.Header.Set("Accept", "application/json, text/plain, */*")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch auth token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch token, status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", err
	}

	token := tokenResp.Token
	if token == "" {
		token = tokenResp.LegacyToken
	}
	if token == "" {
		token = tokenResp.S
	}

	expiresIn := tokenResp.ExpiresIn
	if expiresIn == 0 {
		expiresIn = 600
	}

	s.token = token
	s.tokenExpiresAt = time.Now().Add(time.Duration(TokenTTLMs) * time.Millisecond)

	return s.token, nil
}

func (s *Scraper) authenticatedGet(path string, target interface{}) error {
	token, err := s.getToken()
	if err != nil {
		return err
	}

	separator := "?"
	if strings.Contains(path, "?") {
		separator = "&"
	}
	
	fullURL := fmt.Sprintf("%s%s%s_st=%s", BaseURL, path, separator, token)
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return err
	}

	req.Header.Set("X-Stream-Token", token)
	req.Header.Set("Referer", BaseURL+"/")
	req.Header.Set("Origin", BaseURL)
	req.Header.Set("Accept", "application/json, text/plain, */*")

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed with status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(body, target)
}

func (s *Scraper) GetDetails(tmdbId string) (*scraper.MediaDetails, error) {
	path := fmt.Sprintf("/api/tmdb/details?type=movie&id=%s", tmdbId)
	var tmdbResp TmdbDetailsResponse
	if err := s.authenticatedGet(path, &tmdbResp); err != nil {
		return nil, err
	}

	mediaType := "movie"
	if tmdbResp.NumberOfSeasons > 0 {
		mediaType = "tv"
	}

	var poster, cover string
	if tmdbResp.PosterPath != "" {
		poster = "https://image.tmdb.org/t/p/w500" + tmdbResp.PosterPath
	}
	if tmdbResp.BackdropPath != "" {
		cover = "https://image.tmdb.org/t/p/w1280" + tmdbResp.BackdropPath
	}

	releaseYear := 0
	date := tmdbResp.ReleaseDate
	if date == "" {
		date = tmdbResp.FirstAirDate
	}
	if len(date) >= 4 {
		releaseYear, _ = strconv.Atoi(date[:4])
	}

	title := tmdbResp.Title
	if title == "" {
		title = tmdbResp.Name
	}
	if title == "" {
		title = "Unknown"
	}

	var genres []string
	for _, g := range tmdbResp.Genres {
		genres = append(genres, g.Name)
	}

	details := &scraper.MediaDetails{
		ID:          tmdbId,
		Title:       title,
		Description: tmdbResp.Overview,
		Poster:      poster,
		Cover:       cover,
		Type:        mediaType,
		ReleaseYear: releaseYear,
		Status:      mapStatus(tmdbResp.Status),
		Genres:      genres,
		Rating:      tmdbResp.VoteAverage,
		Episodes:    []scraper.Episode{},
		Sources:     []scraper.StreamSource{},
	}

	return details, nil
}

func mapStatus(status string) string {
	switch strings.ToLower(status) {
	case "returning series":
		return "Ongoing"
	case "ended", "canceled", "released":
		return "Completed"
	default:
		return "Unknown"
	}
}

func (s *Scraper) GetStreamSources(mediaType, tmdbId string, season, episode int) ([]scraper.StreamSource, error) {
	var sources []scraper.StreamSource

	for _, server := range SERVERS {
		var path string
		if mediaType == "tv" {
			path = fmt.Sprintf("%s/id=%s?s=%d&e=%d&type=tv&server=%s", server.Endpoint, tmdbId, season, episode, url.QueryEscape(server.ID))
		} else {
			path = fmt.Sprintf("%s/id=%s?type=movie&server=%s", server.Endpoint, tmdbId, url.QueryEscape(server.ID))
		}

		var streamResp StreamResponse
		err := s.authenticatedGet(path, &streamResp)
		if err != nil {
			log.Printf("1embed server %s failed: %v", server.ID, err)
			continue
		}

		if !streamResp.Success {
			continue
		}

		source := s.parseStreamResponse(&streamResp, server.Name)
		if source != nil {
			sources = append(sources, *source)
		}
	}

	return sources, nil
}

func (s *Scraper) parseStreamResponse(resp *StreamResponse, serverId string) *scraper.StreamSource {
	m3u8Url := resp.Streams.ProxyM3U8
	if m3u8Url == "" {
		m3u8Url = resp.Streams.RawM3U8
	}
	if m3u8Url == "" {
		m3u8Url = resp.Streams.M3U8
	}
	if m3u8Url == "" {
		m3u8Url = resp.StreamURL
	}

	if m3u8Url == "" {
		return nil
	}

	proxiedUrl := fmt.Sprintf("%s/api/proxy/stream.m3u8?url=%s&referer=%s&origin=%s",
		s.apiBaseURL,
		url.QueryEscape(m3u8Url),
		url.QueryEscape(BaseURL+"/"),
		url.QueryEscape(BaseURL),
	)

	var subtitles []scraper.Subtitle
	for _, sub := range resp.Subtitles {
		subUrl := sub.URL
		if subUrl == "" {
			subUrl = sub.File
		}
		if subUrl == "" {
			continue
		}
		lang := sub.Lang
		if lang == "" {
			lang = sub.Label
		}
		if lang == "" {
			lang = "unknown"
		}
		subtitles = append(subtitles, scraper.Subtitle{
			Lang:   lang,
			URL:    subUrl,
			Format: "vtt",
		})
	}

	return &scraper.StreamSource{
		Quality:    "auto",
		URL:        proxiedUrl,
		IsM3U8:     true,
		IsMP4:      false,
		ServerName: serverId,
		Headers: map[string]string{
			"Referer": BaseURL + "/",
			"Origin":  BaseURL,
		},
		Subtitles: subtitles,
	}
}

func (s *Scraper) GetSubtitles(mediaType, tmdbId string, season, episode int) ([]scraper.Subtitle, error) {
	path := fmt.Sprintf("/api/subtitles?type=%s&id=%s&s=%d&e=%d", mediaType, tmdbId, season, episode)
	var subResp SubtitlesResponse
	if err := s.authenticatedGet(path, &subResp); err != nil {
		return nil, err
	}

	var subtitles []scraper.Subtitle
	for _, sub := range subResp.Subtitles {
		subUrl := sub.URL
		if subUrl == "" {
			subUrl = sub.File
		}
		if subUrl == "" {
			continue
		}
		lang := sub.Lang
		if lang == "" {
			lang = sub.Label
		}
		if lang == "" {
			lang = "unknown"
		}
		subtitles = append(subtitles, scraper.Subtitle{
			Lang:   lang,
			URL:    subUrl,
			Format: "vtt",
		})
	}

	return subtitles, nil
}
