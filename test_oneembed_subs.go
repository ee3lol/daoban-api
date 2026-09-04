package main

import (
	"encoding/json"
	"fmt"
	"log"
	
	"ee3lol/daoban-api/internal/scraper/providers/oneembed"
)

func main() {
	scraper := oneembed.NewScraper("http://localhost:8080")
	subs, err := scraper.GetSubtitles("movie", "1386315", 1, 1)
	if err != nil {
		log.Fatal(err)
	}
	
	b, _ := json.MarshalIndent(subs, "", "  ")
	fmt.Println(string(b))
}
