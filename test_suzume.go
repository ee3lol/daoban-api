package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"ee3lol/daoban-api/internal/scraper/providers/vidking"
)

func main() {
	scraper := vidking.NewScraper("http://localhost:8080")
	sources, err := scraper.GetStreamSources("Suzume", "movie", "2022", "603692", "tt16428256", 1, 1)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	
	for _, src := range sources {
		if strings.Contains(src.ServerName, "Yoru") {
			fmt.Printf("Yoru Proxy URL: %s\n", src.URL)
			
			// fetch the m3u8 through proxy? No, the URL is http://localhost:8080/api/proxy/stream.m3u8?url=...
			// Since localhost:8080 might not be running, let's extract the actual URL
			u, _ := strings.CutPrefix(src.URL, "http://localhost:8080/api/proxy/stream.m3u8?url=")
			actualURL := strings.Split(u, "&")[0]
			
			fmt.Println("Actual URL:", actualURL)
			
			// Fetch the master m3u8
			req, _ := http.NewRequest("GET", actualURL, nil)
			req.Header.Set("Origin", "https://vidking.net")
			req.Header.Set("Referer", "https://vidking.net/")
			client := &http.Client{Timeout: 10 * time.Second}
			resp, err := client.Do(req)
			if err != nil {
				fmt.Println("Error fetching m3u8:", err)
				return
			}
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			fmt.Println("Master m3u8:")
			fmt.Println(string(body))
			return
		}
	}
}
