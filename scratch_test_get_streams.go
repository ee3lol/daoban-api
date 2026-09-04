package main

import (
	"encoding/json"
	"fmt"
	"log"

	"ee3lol/daoban-api/internal/scraper/providers/oneembed"
)

func main() {
	scraper := oneembed.NewScraper("http://localhost:8080")
	sources, err := scraper.GetStreamSources("movie", "1355228", 1, 1)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	b, _ := json.MarshalIndent(sources, "", "  ")
	fmt.Println(string(b))
}
