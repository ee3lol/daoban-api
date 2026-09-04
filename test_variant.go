package main

import (
	"fmt"
	"io"
	"net/http"
)

func main() {
	url := "https://cdn.reallyfast.ch/v/v6yMG-pYeF1z1emQ-2dL-SS9Y6riw3zFAa-hF3qg3XGI9Ej5sAba3zKKhQJSQP2gweDkAUL1CVtt_pQpHUDTLOkXGmUgxFnA"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Referer", "https://1embed.cc/")
	req.Header.Set("Origin", "https://1embed.cc")
	req.Header.Set("User-Agent", "Mozilla/5.0")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("Content-Type:", resp.Header.Get("Content-Type"))

	body, _ := io.ReadAll(resp.Body)
	fmt.Println("Body:")
	fmt.Println(string(body)[:500])
}
