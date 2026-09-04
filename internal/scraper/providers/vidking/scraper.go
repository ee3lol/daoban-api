package vidking

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"ee3lol/daoban-api/internal/scraper"
)

const (
	SeedURL       = "https://api.speedracelight.com/seed"
	DecryptAPI    = "https://enc-dec.app/api/dec-videasy"
	OriginReferer = "https://vidking.net"
)

var servers = []struct {
	ID   string
	Name string
}{
	{"cdn", "Bootleg"},
	{"m4uhd", "Siphon"},
	{"vsrc", "Corsair"},
	{"hdmovie", "Outlaw"},
	{"meine", "Smuggler"},
	{"lamovie", "Hijack"},
	{"superflix", "Bandit"},
}

type Scraper struct {
	client     *http.Client
	apiBaseURL string
}

func NewScraper(apiBaseURL string) *Scraper {
	return &Scraper{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		apiBaseURL: apiBaseURL,
	}
}

type SeedResponse struct {
	Seed  string `json:"seed"`
	TTLMs int    `json:"ttlMs"`
}

type DecryptResponse struct {
	Status int `json:"status"`
	Result struct {
		Sources []struct {
			Quality string `json:"quality"`
			URL     string `json:"url"`
		} `json:"sources"`
		Subtitles []struct {
			Lang     string `json:"lang"`
			Language string `json:"language"`
			URL      string `json:"url"`
		} `json:"subtitles"`
		Playlist string `json:"playlist"`
	} `json:"result"`
}

func (s *Scraper) getSeed(tmdbId string) (string, error) {
	reqURL := fmt.Sprintf("%s?mediaId=%s", SeedURL, tmdbId)
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Origin", OriginReferer)
	req.Header.Set("Referer", OriginReferer+"/")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch seed, status: %d", resp.StatusCode)
	}

	var seedResp SeedResponse
	if err := json.NewDecoder(resp.Body).Decode(&seedResp); err != nil {
		return "", err
	}

	return seedResp.Seed, nil
}

type VideasySubtitle struct {
	ID      string `json:"id"`
	URL     string `json:"url"`
	Format  string `json:"format"`
	Display string `json:"display"`
	Lang    string `json:"language"`
}

func (s *Scraper) GetStreamSources(title, mediaType, year, tmdbId, imdbId string, season, episode int) ([]scraper.StreamSource, error) {
	seed, err := s.getSeed(tmdbId)
	if err != nil {
		return nil, fmt.Errorf("failed to get seed: %w", err)
	}

	encTitle := url.QueryEscape(url.QueryEscape(title))
	t := time.Now().UnixMilli()

	// Fetch global subtitles for this IMDb ID
	var globalSubs []scraper.Subtitle
	if imdbId != "" {
		subReqURL := fmt.Sprintf("https://subs.videasy.to/search?id=%s", imdbId)
		subReq, err := http.NewRequest("GET", subReqURL, nil)
		if err == nil {
			subReq.Header.Set("Origin", OriginReferer)
			subReq.Header.Set("Referer", OriginReferer+"/")
			if subResp, err := s.client.Do(subReq); err == nil {
				defer subResp.Body.Close()
				if subResp.StatusCode == 200 {
					var rawSubs []VideasySubtitle
					if err := json.NewDecoder(subResp.Body).Decode(&rawSubs); err == nil {
						for _, rs := range rawSubs {
							// We map it to VTT and use our proxy to convert SRT->VTT
							proxiedSubUrl := fmt.Sprintf("%s/api/proxy/subtitle?url=%s",
								s.apiBaseURL,
								url.QueryEscape(rs.URL),
							)
							globalSubs = append(globalSubs, scraper.Subtitle{
								Lang:   rs.Display,
								URL:    proxiedSubUrl,
								Format: "vtt",
							})
						}
					}
				}
			}
		}
	}

	type ServerResult struct {
		SrvID      string
		SrvName    string
		MaxQuality int
		DecResp    *DecryptResponse
	}

	var results []ServerResult
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, srv := range servers {
		wg.Add(1)
		go func(srvID, srvName string) {
			defer wg.Done()
			
			var reqURL string
			if mediaType == "tv" {
				reqURL = fmt.Sprintf("https://api.speedracelight.com/%s/sources-with-title?title=%s&mediaType=tv&year=%s&episodeId=%d&seasonId=%d&tmdbId=%s&imdbId=%s&enc=2&seed=%s&_t=%d", 
					srvID, encTitle, year, episode, season, tmdbId, imdbId, seed, t)
			} else {
				reqURL = fmt.Sprintf("https://api.speedracelight.com/%s/sources-with-title?title=%s&mediaType=movie&year=%s&episodeId=1&seasonId=1&tmdbId=%s&imdbId=%s&enc=2&seed=%s&_t=%d", 
					srvID, encTitle, year, tmdbId, imdbId, seed, t)
			}

			req, _ := http.NewRequest("GET", reqURL, nil)
			req.Header.Set("Origin", OriginReferer)
			req.Header.Set("Referer", OriginReferer+"/")
			req.Header.Set("Accept", "text/plain")

			resp, err := s.client.Do(req)
			if err != nil {
				return
			}
			defer resp.Body.Close()
			
			if resp.StatusCode != http.StatusOK {
				return
			}

			encData, err := io.ReadAll(resp.Body)
			if err != nil || len(encData) == 0 {
				return
			}

			// Decrypt
			decReqBody := map[string]interface{}{
				"text": string(encData),
				"id":   tmdbId,
				"seed": seed,
			}
			decReqBytes, _ := json.Marshal(decReqBody)

			reqDec, _ := http.NewRequest("POST", DecryptAPI, bytes.NewBuffer(decReqBytes))
			reqDec.Header.Set("Content-Type", "application/json")
			
			respDec, err := s.client.Do(reqDec)
			if err != nil {
				return
			}
			defer respDec.Body.Close()

			var decResp DecryptResponse
			if err := json.NewDecoder(respDec.Body).Decode(&decResp); err != nil {
				return
			}

			if decResp.Status != 200 {
				return
			}

			var maxQuality int
			for _, src := range decResp.Result.Sources {
				qStr := strings.ToLower(src.Quality)
				qStr = strings.ReplaceAll(qStr, "p", "")
				if qStr == "4k" || strings.Contains(qStr, "2160") {
					maxQuality = 2160
				} else {
					q, err := strconv.Atoi(qStr)
					if err == nil && q > maxQuality {
						maxQuality = q
					}
				}
			}
			
			if maxQuality == 0 && (decResp.Result.Playlist != "" || len(decResp.Result.Sources) > 0) {
				maxQuality = 1 
			}

			if maxQuality == 0 {
				fmt.Printf("[%s] Empty: %+v\n", srvName, decResp.Result)
			}

			if maxQuality > 0 {
				mu.Lock()
				results = append(results, ServerResult{
					SrvID:      srvID,
					SrvName:    srvName,
					MaxQuality: maxQuality,
					DecResp:    &decResp,
				})
				mu.Unlock()
			}
		}(srv.ID, srv.Name)
	}

	wg.Wait()

	if len(results) == 0 {
		return nil, nil // No servers worked
	}

	// Sort results by MaxQuality descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].MaxQuality > results[j].MaxQuality
	})

	var sources []scraper.StreamSource

	for _, res := range results {
		var subtitles []scraper.Subtitle
		
		// Extract subtitles
		for _, sub := range res.DecResp.Result.Subtitles {
			lang := sub.Language
			if lang == "" {
				lang = sub.Lang
			}

			proxiedSubUrl := fmt.Sprintf("%s/api/proxy/stream.vtt?url=%s&referer=%s&origin=%s",
				s.apiBaseURL,
				url.QueryEscape(sub.URL),
				url.QueryEscape(OriginReferer+"/"),
				url.QueryEscape(OriginReferer),
			)

			subtitles = append(subtitles, scraper.Subtitle{
				Lang:   lang,
				URL:    proxiedSubUrl,
				Format: "vtt",
			})
		}
		
		// Append global videasy subtitles
		subtitles = append(subtitles, globalSubs...)

		masterUrl := res.DecResp.Result.Playlist
		var audioTracks []string

		if len(res.DecResp.Result.Sources) > 0 {
			firstSource := res.DecResp.Result.Sources[0].URL
			if masterUrl == "" {
				masterUrl = firstSource
			}
			
			// Probe for audio tracks
			if strings.Contains(firstSource, "-a1.m3u8") {
				audioTracks = append(audioTracks, "1")
				
				var probeWg sync.WaitGroup
				var probeMu sync.Mutex
				for i := 2; i <= 5; i++ {
					probeWg.Add(1)
					go func(trackNum int) {
						defer probeWg.Done()
						testUrl := strings.Replace(firstSource, "-a1.m3u8", fmt.Sprintf("-a%d.m3u8", trackNum), 1)
						req, _ := http.NewRequest("HEAD", testUrl, nil)
						req.Header.Set("Origin", OriginReferer)
						req.Header.Set("Referer", OriginReferer+"/")
						
						client := &http.Client{Timeout: 3 * time.Second}
						resp, err := client.Do(req)
						if err == nil {
							defer resp.Body.Close()
							if resp.StatusCode == 200 {
								probeMu.Lock()
								audioTracks = append(audioTracks, fmt.Sprintf("%d", trackNum))
								probeMu.Unlock()
							}
						}
					}(i)
				}
				probeWg.Wait()
				
				sort.Slice(audioTracks, func(i, j int) bool {
					return audioTracks[i] < audioTracks[j]
				})
			}
		}

		if masterUrl != "" {
			proxiedUrl := fmt.Sprintf("%s/api/proxy/stream.m3u8?url=%s&referer=%s&origin=%s",
				s.apiBaseURL,
				url.QueryEscape(masterUrl),
				url.QueryEscape(OriginReferer+"/"),
				url.QueryEscape(OriginReferer),
			)

			sources = append(sources, scraper.StreamSource{
				Quality:     "auto",
				URL:         proxiedUrl,
				IsM3U8:      true,
				IsMP4:       false,
				ServerName:  res.SrvName,
				AudioTracks: audioTracks,
				Headers: map[string]string{
					"Referer": OriginReferer + "/",
					"Origin":  OriginReferer,
				},
				Subtitles: subtitles,
			})
		}
	}

	return sources, nil
}
