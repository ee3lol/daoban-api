package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

func main() {
	tmdbId := "1078605"
	title := "Weapons"
	encTitle := url.QueryEscape(url.QueryEscape(title))
	
	seedUrl := fmt.Sprintf("https://api.speedracelight.com/seed?mediaId=%s", tmdbId)
	req, _ := http.NewRequest("GET", seedUrl, nil)
	req.Header.Set("Origin", "https://vidking.net")
	req.Header.Set("Referer", "https://vidking.net/")
	
	resp, _ := http.DefaultClient.Do(req)
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	
	var seedResp map[string]interface{}
	json.Unmarshal(body, &seedResp)
	seed := seedResp["seed"].(string)
	
	t := time.Now().UnixMilli()
	sourcesUrl := fmt.Sprintf("https://api.speedracelight.com/m4uhd/sources-with-title?title=%s&mediaType=movie&year=2025&episodeId=1&seasonId=1&tmdbId=%s&imdbId=tt26581740&enc=2&seed=%s&_t=%d", encTitle, tmdbId, seed, t)
	
	req2, _ := http.NewRequest("GET", sourcesUrl, nil)
	req2.Header.Set("Origin", "https://vidking.net")
	req2.Header.Set("Referer", "https://vidking.net/")
	req2.Header.Set("Accept", "text/plain")
	
	resp2, _ := http.DefaultClient.Do(req2)
	encData, _ := io.ReadAll(resp2.Body)
	resp2.Body.Close()
	
	// decrypt
	decReqBody := map[string]interface{}{
		"text": string(encData),
		"id": tmdbId,
		"seed": seed,
	}
	decReqBytes, _ := json.Marshal(decReqBody)
	
	req3, _ := http.NewRequest("POST", "https://enc-dec.app/api/dec-videasy", bytes.NewBuffer(decReqBytes))
	req3.Header.Set("Content-Type", "application/json")
	resp3, _ := http.DefaultClient.Do(req3)
	decData, _ := io.ReadAll(resp3.Body)
	resp3.Body.Close()
	
	fmt.Printf("Decrypted Data: %s\n", string(decData))
}
