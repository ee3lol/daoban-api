package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type VideasySubtitle struct {
	ID      string `json:"id"`
	URL     string `json:"url"`
	Format  string `json:"format"`
	Display string `json:"display"`
	Lang    string `json:"language"`
}

func main() {
	req, _ := http.NewRequest("GET", "https://subs.videasy.to/search?id=tt26581740", nil)
	req.Header.Set("Origin", "https://vidking.net")
	req.Header.Set("Referer", "https://vidking.net/")
	
	resp, _ := http.DefaultClient.Do(req)
	body, _ := io.ReadAll(resp.Body)
	
	var subs []VideasySubtitle
	json.Unmarshal(body, &subs)
	
	fmt.Printf("Fetched %d subtitles!\n", len(subs))
	if len(subs) > 0 {
		fmt.Printf("First sub: %s - %s\n", subs[0].Display, subs[0].URL)
	}
}
