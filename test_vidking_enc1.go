package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type SeedResponse struct {
	Seed string `json:"seed"`
}

func main() {
	tmdbId := "1078605"
	
	seedUrl := fmt.Sprintf("https://api.speedracelight.com/seed?mediaId=%s", tmdbId)
	req, _ := http.NewRequest("GET", seedUrl, nil)
	req.Header.Set("Origin", "https://vidking.net")
	req.Header.Set("Referer", "https://vidking.net/")
	
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Error getting seed: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	var seedResp SeedResponse
	json.Unmarshal(body, &seedResp)
	
	seed := seedResp.Seed
	
	t := time.Now().UnixMilli()
	sourcesUrl := fmt.Sprintf("https://api.speedracelight.com/cdn/sources-with-title?title=Weapons&mediaType=movie&year=2025&episodeId=1&seasonId=1&tmdbId=%s&imdbId=tt26581740&enc=1&seed=%s&_t=%d", tmdbId, seed, t) // enc=1
	
	req2, _ := http.NewRequest("GET", sourcesUrl, nil)
	req2.Header.Set("Origin", "https://vidking.net")
	req2.Header.Set("Referer", "https://vidking.net/")
	
	resp2, err := http.DefaultClient.Do(req2)
	if err != nil {
		fmt.Printf("Error getting sources: %v\n", err)
		return
	}
	defer resp2.Body.Close()
	
	body2, _ := io.ReadAll(resp2.Body)
	fmt.Printf("Sources Response (enc=1): %s\n", string(body2))
}
