package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"ee3lol/daoban-api/internal/scraper/providers/oneembed"
)

func main() {
	scraper := oneembed.NewScraper("http://localhost:8080")
	fmt.Println("Fetching stream sources for TV 125988 (Silo, S1E1)")

	sources, err := scraper.GetStreamSources("tv", "125988", 1, 1)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	var outlawUrl string
	for _, src := range sources {
		if src.ServerName == "Outlaw [4k]" || src.ServerName == "Outlaw" || strings.Contains(src.ServerName, "Outlaw") {
			outlawUrl = src.URL
			fmt.Printf("Found Outlaw server (%s), URL: %s\n", src.ServerName, src.URL)
			break
		}
	}

	if outlawUrl == "" {
		fmt.Println("Outlaw server not found in response!")
		fmt.Println("Servers found:")
		for _, s := range sources {
			fmt.Printf(" - %s\n", s.ServerName)
		}
		return
	}

	parsed, _ := url.Parse(outlawUrl)
	targetUrl := parsed.Query().Get("url")
	if targetUrl == "" {
		targetUrl = outlawUrl
	}

	fmt.Printf("\nFetching M3U8 from: %s\n", targetUrl)
	req, _ := http.NewRequest("GET", targetUrl, nil)
	req.Header.Set("Referer", "https://1embed.cc/")
	req.Header.Set("Origin", "https://1embed.cc")
	req.Header.Set("User-Agent", "Mozilla/5.0")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Failed to fetch M3U8:", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	m3u8Content := string(body)

	fmt.Println("\nM3U8 Content:")
	fmt.Println(m3u8Content)

	if strings.Contains(m3u8Content, "RESOLUTION=3840x2160") || strings.Contains(m3u8Content, "2160p") || strings.Contains(m3u8Content, "4k") || strings.Contains(m3u8Content, "4K") {
		fmt.Println("\n✅ YES, WE ARE GETTING 4K ON OUTLAW FOR 125988!")
	} else {
		fmt.Println("\n❌ NO 4K FOUND IN M3U8!")
	}
}
