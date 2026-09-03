package anidb

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"ee3lol/daoban-api/internal/scraper"
)

type Scraper struct {
	client     *http.Client
	apiBaseURL string
}

func NewScraper(apiBaseURL string) *Scraper {
	return &Scraper{
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
		apiBaseURL: apiBaseURL,
	}
}

func (s *Scraper) GetStreamSourcesForAnime(title string, season, episode int) ([]scraper.StreamSource, error) {
	// 1. Search for the anime ID
	searchURL := fmt.Sprintf("https://anidb.app/search/suggestions?q=%s", url.QueryEscape(title))
	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to search anidb: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	bodyStr := string(bodyBytes)

	// Extract the ID from the first href. E.g. href="https://anidb.app/anime/chainsaw-man-922"
	re := regexp.MustCompile(`href="https://anidb\.app/anime/[^"]+-(\d+)"`)
	matches := re.FindStringSubmatch(bodyStr)
	if len(matches) < 2 {
		return nil, fmt.Errorf("could not find anime ID for %s", title)
	}
	animeID := matches[1]

	// 2. Fetch episodes list
	episodesURL := fmt.Sprintf("https://anidb.app/api/frontend/anime/%s/episodes", animeID)
	req, err = http.NewRequest("GET", episodesURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Referer", fmt.Sprintf("https://anidb.app/anime/something-%s", animeID))

	resp, err = s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch episodes: %w", err)
	}
	defer resp.Body.Close()

	var epResp EpisodesResponse
	if err := json.NewDecoder(resp.Body).Decode(&epResp); err != nil {
		return nil, fmt.Errorf("failed to parse episodes: %w", err)
	}

	var targetEpID int = -1
	for _, ep := range epResp.Episodes {
		if ep.Number == episode {
			targetEpID = ep.ID
			break
		}
	}

	if targetEpID == -1 {
		return nil, fmt.Errorf("episode %d not found", episode)
	}

	// 3. Fetch languages/embeds for the episode
	langsURL := fmt.Sprintf("https://anidb.app/api/frontend/episode/%d/languages", targetEpID)
	req, err = http.NewRequest("GET", langsURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Referer", fmt.Sprintf("https://anidb.app/anime/something-%s", animeID))

	resp, err = s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch languages: %w", err)
	}
	defer resp.Body.Close()

	var langResp LanguagesResponse
	if err := json.NewDecoder(resp.Body).Decode(&langResp); err != nil {
		return nil, fmt.Errorf("failed to parse languages: %w", err)
	}

	var sources []scraper.StreamSource
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Extract streams in parallel
	for _, lang := range langResp.Languages {
		wg.Add(1)
		go func(l Language) {
			defer wg.Done()
			
			// We only care about English (dub) and Japanese (sub)
			var audioType string
			if l.Code == "eng" {
				audioType = "dub"
			} else if l.Code == "jpn" {
				audioType = "sub"
			} else {
				return // skip other languages
			}

			streamURL := s.extractStreamFromEmbed(l.EmbedURL)
			if streamURL != "" {
				proxiedUrl := fmt.Sprintf("%s/api/proxy/stream.m3u8?url=%s&referer=%s&origin=%s",
					s.apiBaseURL,
					url.QueryEscape(streamURL),
					url.QueryEscape("https://anidb.app/"),
					url.QueryEscape("https://anidb.app"),
				)

				mu.Lock()
				sources = append(sources, scraper.StreamSource{
					Quality:    "auto",
					URL:        proxiedUrl,
					IsM3U8:     true,
					IsMP4:      false,
					IsEmbed:    false,
					ServerName: "Bootleg",
					AudioType:  audioType,
					Headers: map[string]string{
						"Origin":  "https://anidb.app",
						"Referer": "https://anidb.app/",
					},
					Subtitles: []scraper.Subtitle{},
				})
				mu.Unlock()
			}
		}(lang)
	}

	wg.Wait()
	return sources, nil
}

func (s *Scraper) extractStreamFromEmbed(embedURL string) string {
	req, err := http.NewRequest("GET", embedURL, nil)
	if err != nil {
		return ""
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
	req.Header.Set("Referer", "https://anidb.app/")

	resp, err := s.client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	html := string(bodyBytes)

	// Look for: file: 'https://hls.anidb.app/stream/.../master.m3u8'
	re := regexp.MustCompile(`file:\s*'([^']+)'`)
	matches := re.FindStringSubmatch(html)
	if len(matches) >= 2 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}
