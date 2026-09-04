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
	// Let's use Deadpool & Wolverine TMDB ID: 533535
	sources, err := scraper.GetStreamSources("Deadpool & Wolverine", "movie", "2024", "533535", "tt6263850", 1, 1)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	
	if len(sources) == 0 {
		fmt.Println("No sources found")
		return
	}
	
	for _, src := range sources {
		if strings.Contains(src.ServerName, "Yoru") {
			fmt.Printf("Yoru Master URL: %s\n", src.URL)
			return
		}
	}
}
