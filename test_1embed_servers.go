package main

import (
	"fmt"
	"ee3lol/daoban-api/internal/scraper/providers/oneembed"
)

func main() {
	s := oneembed.NewScraper("http://localhost:3001")
	
	// Create temporary servers list to test
	servers := []oneembed.ServerConfig{
		{ID: "NORE", Name: "Nore", Endpoint: "/api/sources/2"},
		{ID: "GORE", Name: "Gore", Endpoint: "/api/sources/3"},
		{ID: "ZORE", Name: "Zore [HD]", Endpoint: "/api/sources/1"},
		{ID: "BORE", Name: "Bore [4K]", Endpoint: "/api/sources/4"},
	}

	for _, srv := range servers {
		fmt.Printf("Testing %s...\n", srv.Name)
		var streamResp oneembed.StreamResponse
		// Using an internal method is hard from main without exposing it, so let's just make the HTTP request directly
		// Wait, I can't easily call authenticatedGet since it's unexported. 
	}
}
