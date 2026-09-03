package main

import (
	"fmt"
	"encoding/json"
	"ee3lol/daoban-api/internal/scraper/providers/oneembed"
)

func main() {
	scraper := oneembed.NewScraper("http://localhost:3001")
	sources, err := scraper.GetStreamSources("movie", "1386315", 1, 1)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	for _, src := range sources {
		if src.ServerName == "Outlaw [4k]" {
			b, _ := json.MarshalIndent(src, "", "  ")
			fmt.Println(string(b))
		}
	}
}
