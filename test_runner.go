package main

import (
	"encoding/json"
	"fmt"
	"log"
	
	"ee3lol/daoban-api/internal/scraper/providers/vidking"
)

func main() {
	scraper := vidking.NewScraper("http://localhost:8080")
	sources, err := scraper.GetStreamSources("The Runner", "movie", "2024", "1386315", "tt29731674", 1, 1)
	if err != nil {
		log.Fatal(err)
	}
	
	b, _ := json.MarshalIndent(sources, "", "  ")
	fmt.Println(string(b))
}
