package main

import (
	"encoding/json"
	"fmt"
	"log"
	
	"ee3lol/daoban-api/internal/scraper"
	"ee3lol/daoban-api/internal/scraper/providers/oneembed"
	"ee3lol/daoban-api/internal/scraper/providers/vidking"
)

func main() {
	vk := vidking.NewScraper("http://localhost:3001")
	oe := oneembed.NewScraper("http://localhost:3001")
	
	fmt.Println("Fetching Yoru sources...")
	sources, _ := vk.GetStreamSources("The Runner", "movie", "2024", "1386315", "tt29731674", 1, 1)
	
	fmt.Println("Fetching global subtitles...")
	subs, err := oe.GetSubtitles("movie", "1386315", 1, 1)
	if err != nil {
		log.Fatal(err)
	}
	
	// Merge
	for i := range sources {
		// Just append all global subs to the source's subtitles
		sources[i].Subtitles = append(sources[i].Subtitles, subs...)
	}
	
	b, _ := json.MarshalIndent(sources[0].Subtitles, "", "  ")
	fmt.Println(string(b))
}
