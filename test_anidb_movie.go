package main

import (
	"ee3lol/daoban-api/internal/scraper/providers/anidb"
	"fmt"
)

func main() {
	scraper := anidb.NewScraper("")
	sources, err := scraper.GetStreamSourcesForAnime("Spirited Away", 1, 1)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	fmt.Printf("Sources: %+v\n", sources)
}
