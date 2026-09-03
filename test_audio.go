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
	for retries := 0; retries < 5; retries++ {
		sources, _ = scraper.GetStreamSources("movie", "1386315", 1, 1)
		for _, src := range sources {
			if src.ServerName == "Outlaw [4k]" {
				outlawUrl = src.URL
				break
			}
		}
		if outlawUrl != "" {
			break
		}
		time.Sleep(2 * time.Second)
	}
	
	if outlawUrl == "" {
		fmt.Println("Still not found.")
		return
	}

	parsed, _ := url.Parse(outlawUrl)
	q := parsed.Query()
	targetUrl := q.Get("url")
	
	client := &http.Client{Timeout: 15 * time.Second}
	req, _ := http.NewRequest("GET", targetUrl, nil)
	if origin := q.Get("origin"); origin != "" {
		req.Header.Set("Origin", origin)
	}
	if referer := q.Get("referer"); referer != "" {
		req.Header.Set("Referer", referer)
	}
	if ua := q.Get("userAgent"); ua != "" {
		req.Header.Set("User-Agent", ua)
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error fetching M3U8: %v\n", err)
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	playlist := string(body)

	scanner := bufio.NewScanner(strings.NewReader(playlist))
	fmt.Println("\n--- Audio Tracks in M3U8 ---")
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#EXT-X-MEDIA:TYPE=AUDIO") {
			fmt.Println(line)
		}
	}
}
