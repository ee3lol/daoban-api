package main

import (
	"fmt"
	"io"
	"net/http"
)

func main() {
	target := "https://shrill-shape-1d18.vr6q4oelwfusw.workers.dev/p/afc7d47f/ma9yIsUHLd1oEUKmveBRD6CL9aSh9fi08J7f9fAYJaJBpCdYD0B-GAdVd7Zg6lACo8F67V8n20Xhl51_593TUkhcd9hFZcoR7LlT9DOAAq0slTZjhKqJapt6Dq7jqdYGP_T1DrwXxdNPlUHYHMW7zmEB27NpDMwBlzcFeaWS7zrVG6LCHKS49JF1Tk1djkqL4MUFY-vDHWSKhELeb8b37sTdNUWrEn_y8OkbL86LfL_-a9lfOZyclfrk_pmiS4FqVJF5O5-fteri7BJKVO_H_ekZNZjrMZm60VywmqFz41O9Mt33qAjhHR6t7zRsEwvjcgSE4ODPk1Lt3usInTACtg.m3u8"
	
	req, _ := http.NewRequest("GET", target, nil)
	req.Header.Set("Origin", "https://vidking.net")
	req.Header.Set("Referer", "https://vidking.net/")
	
	resp, _ := http.DefaultClient.Do(req)
	fmt.Printf("Status with vidking: %d\n", resp.StatusCode)
	
	req2, _ := http.NewRequest("GET", target, nil)
	req2.Header.Set("Origin", "https://player.videasy.to")
	req2.Header.Set("Referer", "https://player.videasy.to/")
	
	resp2, _ := http.DefaultClient.Do(req2)
	fmt.Printf("Status with videasy: %d\n", resp2.StatusCode)
	
	if resp2.StatusCode == 200 {
		b, _ := io.ReadAll(resp2.Body)
		fmt.Printf("Body starts with: %s\n", string(b[:50]))
	}
}
