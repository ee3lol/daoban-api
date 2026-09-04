package main

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
)

func main() {
	url := "https://example.com/playlist.m3u8" // Will test the logic with a sample text
	
	sampleM3u8 := `#EXTM3U
#EXT-X-STREAM-INF:BANDWIDTH=5400000,RESOLUTION=1920x1080
1080/index.m3u8
#EXT-X-STREAM-INF:BANDWIDTH=2400000,RESOLUTION=1280x720
720/index.m3u8`

	re := regexp.MustCompile(`RESOLUTION=\d+x(\d+)`)
	matches := re.FindAllStringSubmatch(sampleM3u8, -1)
	maxRes := 0
	for _, match := range matches {
		if len(match) > 1 {
			res, _ := strconv.Atoi(match[1])
			if res > maxRes {
				maxRes = res
			}
		}
	}
	
	fmt.Printf("Max Resolution: %d\n", maxRes)
}
