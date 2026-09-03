package main

import (
	"fmt"
	"ee3lol/daoban-api/internal/scraper/providers/anidb"
)

func main() {
	scraper := anidb.NewScraper("")
	sources, err := scraper.GetStreamSourcesForAnime("One Piece", 1, 1)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	fmt.Printf("Sources: %+v\n", sources)
}
