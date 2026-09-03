package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"ee3lol/daoban-api/internal/scraper/providers/oneembed"
)

func main() {
	scraper := oneembed.NewScraper("http://localhost:3001")
	fmt.Println("Fetching sources from oneembed...")
	sources, err := scraper.GetStreamSources("movie", "1386315", 1, 1)
	if err != nil {
		fmt.Printf("Error getting sources: %v\n", err)
		return
	}

	var outlawUrl string
	for _, src := range sources {
		if src.ServerName == "Outlaw [4k]" {
			outlawUrl = src.URL
			break
		}
	}

	if outlawUrl == "" {
		fmt.Println("Outlaw [4k] server not found in response!")
		return
	}

	fmt.Printf("Found Outlaw [4k] Proxy URL:\n%s\n\n", outlawUrl)

	// Extract the actual CDN URL and headers from the proxy URL
	parsed, err := url.Parse(outlawUrl)
	if err != nil {
		fmt.Printf("Error parsing proxy URL: %v\n", err)
		return
	}

	q := parsed.Query()
	targetUrl := q.Get("url")
	origin := q.Get("origin")
	referer := q.Get("referer")
	userAgent := q.Get("userAgent")

	if targetUrl == "" {
		fmt.Println("Could not extract target URL from proxy URL")
		return
	}

	fmt.Printf("Target CDN URL: %s\n", targetUrl)
	fmt.Printf("Using Headers:\n - Origin: %s\n - Referer: %s\n - User-Agent: %s\n\n", origin, referer, userAgent)

	// Fetch the actual M3U8 playlist
	client := &http.Client{Timeout: 15 * time.Second}
	req, _ := http.NewRequest("GET", targetUrl, nil)
	
	if origin != "" {
		req.Header.Set("Origin", origin)
	}
	if referer != "" {
		req.Header.Set("Referer", referer)
	}
	if userAgent != "" {
		req.Header.Set("User-Agent", userAgent)
	}

	fmt.Println("Fetching master M3U8 playlist...")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error fetching M3U8: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Failed to fetch M3U8, HTTP Status: %d\n", resp.StatusCode)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		return
	}

	playlist := string(body)
	scanner := bufio.NewScanner(strings.NewReader(playlist))
	
	var resolutions []string
	var bandwidths []string
	is4k := false

	fmt.Println("\n--- M3U8 Stream Info ---")
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#EXT-X-STREAM-INF:") {
			// Extract RESOLUTION and BANDWIDTH
			parts := strings.Split(line, ",")
			var res, bw string
			for _, part := range parts {
				if strings.Contains(part, "RESOLUTION=") {
					res = strings.Split(part, "=")[1]
					resolutions = append(resolutions, res)
					if res == "3840x2160" || res == "4096x2160" {
						is4k = true
					}
				}
				if strings.Contains(part, "BANDWIDTH=") {
					bw = strings.Split(part, "=")[1]
					bandwidths = append(bandwidths, bw)
				}
			}
			fmt.Printf("Stream Found - Resolution: %-12s Bandwidth: %s\n", res, bw)
		}
	}

	fmt.Println("------------------------\n")
	
	if len(resolutions) == 0 {
		fmt.Println("No resolutions found in the playlist. It might be a single-stream or chunklist M3U8.")
	} else {
		if is4k {
			fmt.Println("✅ RESULT: TRUE 4K STREAM DETECTED! (3840x2160 or 4096x2160)")
		} else {
			highestRes := resolutions[len(resolutions)-1]
			fmt.Printf("❌ RESULT: Not 4K. Highest resolution found is %s\n", highestRes)
		}
	}
}
