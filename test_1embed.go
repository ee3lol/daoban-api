package main

import (
	"fmt"
	"ee3lol/daoban-api/internal/scraper/providers/oneembed"
)

func main() {
	s := oneembed.NewScraper("http://localhost:3001")
	sources, _ := s.GetStreamSources("movie", "550", 1, 1)
	fmt.Printf("%+v\n", sources)
}
