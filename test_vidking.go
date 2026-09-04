package main

import (
	"fmt"
	"ee3lol/daoban-api/internal/scraper/providers/vidking"
)

func main() {
	scraper := vidking.NewScraper("http://localhost:3000")
	sources, err := scraper.GetStreamSources("Iron Man", "movie", "2008", "1726", "tt0371746", 1, 1)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Printf("Found %d sources:\n", len(sources))
	for _, s := range sources {
		fmt.Println(s.ServerName)
	}
}
