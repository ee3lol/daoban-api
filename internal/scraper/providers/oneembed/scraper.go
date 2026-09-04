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
			Timeout: 30 * time.Second,
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
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36")

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
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36")

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

func (s *Scraper) GetDetails(mediaType, tmdbId string) (*scraper.MediaDetails, error) {
	path := fmt.Sprintf("/api/tmdb/details?type=%s&id=%s", mediaType, tmdbId)
	var tmdbResp TmdbDetailsResponse
	if err := s.authenticatedGet(path, &tmdbResp); err != nil {
		return nil, err
	}

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
	sourceChan := make(chan *scraper.StreamSource, len(SERVERS))

	for _, server := range SERVERS {
		go func(srv ServerConfig) {
			var path string
			if mediaType == "tv" {
				path = fmt.Sprintf("%s/id=%s?s=%d&e=%d&type=tv&server=%s", srv.Endpoint, tmdbId, season, episode, url.QueryEscape(srv.ID))
			} else {
				path = fmt.Sprintf("%s/id=%s?type=movie&server=%s", srv.Endpoint, tmdbId, url.QueryEscape(srv.ID))
			}

			var streamResp StreamResponse
			err := s.authenticatedGet(path, &streamResp)
			if err != nil {
				log.Printf("1embed server %s failed: %v", srv.ID, err)
				sourceChan <- nil
				return
			}

			if !streamResp.Success {
				sourceChan <- nil
				return
			}

			sourceChan <- s.parseStreamResponse(&streamResp, srv.Name)
		}(server)
	}

	timeout := time.After(25 * time.Second)
	for i := 0; i < len(SERVERS); i++ {
		select {
		case src := <-sourceChan:
			if src != nil {
				sources = append(sources, *src)
			}
		case <-timeout:
			log.Printf("GetStreamSources hit 25s timeout, returning %d sources found so far", len(sources))
			return sources, nil
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

	origin := BaseURL
	referer := BaseURL + "/"
	userAgent := ""

	if strings.Contains(m3u8Url, "proxy.lalis.lol") {
		parsedProxy, err := url.Parse(m3u8Url)
		if err == nil {
			targetUrl := parsedProxy.Query().Get("url")
			if targetUrl != "" {
				m3u8Url = targetUrl
			}
			headersJSON := parsedProxy.Query().Get("headers")
			if headersJSON != "" {
				var headersMap map[string]string
				if err := json.Unmarshal([]byte(headersJSON), &headersMap); err == nil {
					if orig, ok := headersMap["Origin"]; ok && orig != "" {
						origin = orig
					}
					if ref, ok := headersMap["Referer"]; ok && ref != "" {
						referer = ref
					}
					if ua, ok := headersMap["User-Agent"]; ok && ua != "" {
						userAgent = ua
					}
				}
			}
		}
	}

	proxiedUrl := fmt.Sprintf("%s/api/proxy/stream.m3u8?url=%s&referer=%s&origin=%s",
		s.apiBaseURL,
		url.QueryEscape(m3u8Url),
		url.QueryEscape(referer),
		url.QueryEscape(origin),
	)
	if userAgent != "" {
		proxiedUrl += fmt.Sprintf("&userAgent=%s", url.QueryEscape(userAgent))
	}

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
