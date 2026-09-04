package main

import (
	"fmt"
	"io"
	"net/http"
)

func main() {
	target := "https://moon.peakstorm.top/vd/NmRaUEN4RlpsZS1scmNWS21BVVczUTpXUWVMOURVUXRnRU9iMEdpdThMOFh2VWEwTk5nR0hFQ0NNTFpxSW51ck13/master.m3u8"
	
	req, _ := http.NewRequest("GET", target, nil)
	req.Header.Set("Origin", "https://player.videasy.to")
	req.Header.Set("Referer", "https://player.videasy.to/")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
	
	resp, _ := http.DefaultClient.Do(req)
	fmt.Printf("Status with videasy: %d\n", resp.StatusCode)
	
	req2, _ := http.NewRequest("GET", target, nil)
	req2.Header.Set("Origin", "https://vidking.net")
	req2.Header.Set("Referer", "https://vidking.net/")
	req2.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
	
	resp2, _ := http.DefaultClient.Do(req2)
	fmt.Printf("Status with vidking: %d\n", resp2.StatusCode)
	
	b, _ := io.ReadAll(resp2.Body)
	fmt.Printf("Body: %s\n", string(b))
}
