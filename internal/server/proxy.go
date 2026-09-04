package server

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

func setupProxyRoutes(mux *http.ServeMux) {
	proxyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		targetUrl := r.URL.Query().Get("url")
		if targetUrl == "" {
			http.Error(w, "url parameter is required", http.StatusBadRequest)
			return
		}

		// HOTLINK & SCRAPER PROTECTION
		// If the request comes from an origin/referer that isn't ours, trap it!
		reqOrigin := r.Header.Get("Origin")
		reqReferer := r.Header.Get("Referer")
		
		isHotlink := false
		if reqOrigin != "" && !strings.Contains(reqOrigin, "daoban.lol") && !strings.Contains(reqOrigin, "localhost") {
			isHotlink = true
		}
		if reqOrigin == "" && reqReferer != "" && !strings.Contains(reqReferer, "daoban.lol") && !strings.Contains(reqReferer, "localhost") {
			isHotlink = true
		}

		if isHotlink {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.WriteHeader(http.StatusForbidden)
			
			// THE TAR PIT: Waste their time by holding the connection open
			// We send one space character every 10 seconds for 5 minutes.
			// This completely breaks their video player and ties up their proxy servers.
			for i := 0; i < 30; i++ {
				w.Write([]byte(" "))
				if flusher, ok := w.(http.Flusher); ok {
					flusher.Flush()
				}
				time.Sleep(10 * time.Second)
			}
			w.Write([]byte(`{"error": "bro just dm admins on discord for the src code its free lol"}`))
			return
		}

		cookie := r.URL.Query().Get("cookie")
		referer := r.URL.Query().Get("referer")
		origin := r.URL.Query().Get("origin")
		userAgent := r.URL.Query().Get("userAgent")
		audioTrack := r.URL.Query().Get("audioTrack")

		if origin == "" {
			origin = "https://ee3.me"
		}
		if referer == "" {
			referer = "https://ee3.me/"
		}
		if userAgent == "" {
			userAgent = r.Header.Get("User-Agent")
		}
		if userAgent == "" {
			userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36"
		}
		
		// If audioTrack is specified, rewrite the target URL to point to that track
		if audioTrack != "" && audioTrack != "1" {
			targetUrl = strings.Replace(targetUrl, "-a1.m3u8", fmt.Sprintf("-a%s.m3u8", audioTrack), 1)
			targetUrl = strings.Replace(targetUrl, "-a1.mp4", fmt.Sprintf("-a%s.mp4", audioTrack), 1)
			targetUrl = strings.Replace(targetUrl, "-a1.m4s", fmt.Sprintf("-a%s.m4s", audioTrack), 1)
		}

		req, err := http.NewRequest("GET", targetUrl, nil)
		if err != nil {
			log.Printf("Proxy error creating request for %s: %v", targetUrl, err)
			http.Error(w, `{"success":false,"error":"Proxy fetch failed"}`, http.StatusInternalServerError)
			return
		}

		req.Header.Set("Origin", origin)
		req.Header.Set("Referer", referer)
		req.Header.Set("User-Agent", userAgent)
		if cookie != "" {
			req.Header.Set("Cookie", cookie)
		}
		if rangeHdr := r.Header.Get("Range"); rangeHdr != "" {
			req.Header.Set("Range", rangeHdr)
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Proxy error fetching %s: %v", targetUrl, err)
			http.Error(w, `{"success":false,"error":"Proxy fetch failed"}`, http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		contentType := strings.ToLower(resp.Header.Get("Content-Type"))
		isM3U8 := strings.Contains(contentType, "mpegurl") || strings.Contains(contentType, "m3u8") || 
			(strings.Contains(targetUrl, ".m3u8") && !strings.Contains(targetUrl, ".html") && !strings.Contains(targetUrl, ".ts"))

		if isM3U8 {
			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				http.Error(w, "Failed to read m3u8 body", http.StatusInternalServerError)
				return
			}
			playlist := string(bodyBytes)

			parsedUrl, _ := url.Parse(targetUrl)
			basePath := parsedUrl.Path[:strings.LastIndex(parsedUrl.Path, "/")+1]
			baseUrlOrigin := parsedUrl.Scheme + "://" + parsedUrl.Host

			var rewrittenLines []string
			lines := strings.Split(playlist, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" {
					rewrittenLines = append(rewrittenLines, line)
					continue
				}

				buildProxyUrl := func(target string) string {
					segmentUrl := target
					if !strings.HasPrefix(segmentUrl, "http") {
						if strings.HasPrefix(segmentUrl, "/") {
							segmentUrl = baseUrlOrigin + segmentUrl
						} else {
							segmentUrl = baseUrlOrigin + basePath + segmentUrl
						}
					}

					proxyUrl := fmt.Sprintf("/api/proxy/stream.m3u8?url=%s", url.QueryEscape(segmentUrl))
					if cookie != "" {
						proxyUrl += fmt.Sprintf("&cookie=%s", url.QueryEscape(cookie))
					}
					if referer != "" {
						proxyUrl += fmt.Sprintf("&referer=%s", url.QueryEscape(referer))
					}
					if origin != "" {
						proxyUrl += fmt.Sprintf("&origin=%s", url.QueryEscape(origin))
					}
					if audioTrack != "" {
						proxyUrl += fmt.Sprintf("&audioTrack=%s", url.QueryEscape(audioTrack))
					}
					return proxyUrl
				}

				if strings.HasPrefix(line, "#") {
					// We must rewrite URIs inside tags like #EXT-X-MEDIA and #EXT-X-MAP
					if strings.Contains(line, `URI="`) {
						parts := strings.Split(line, `URI="`)
						if len(parts) == 2 {
							beforeUri := parts[0]
							afterUriStart := parts[1]
							endQuoteIdx := strings.Index(afterUriStart, `"`)
							if endQuoteIdx != -1 {
								originalUri := afterUriStart[:endQuoteIdx]
								afterUriEnd := afterUriStart[endQuoteIdx:]
								line = beforeUri + `URI="` + buildProxyUrl(originalUri) + afterUriEnd
							}
						}
					}
					rewrittenLines = append(rewrittenLines, line)
					continue
				}

				rewrittenLines = append(rewrittenLines, buildProxyUrl(line))
			}

			rewrittenPlaylist := strings.Join(rewrittenLines, "\n")

			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
			w.Header().Set("Content-Length", fmt.Sprintf("%d", len(rewrittenPlaylist)))
			w.Header().Set("Cache-Control", "public, max-age=10")
			w.WriteHeader(resp.StatusCode)
			w.Write([]byte(rewrittenPlaylist))
			return
		}

		for k, vv := range resp.Header {
			kLower := strings.ToLower(k)
			if kLower == "transfer-encoding" || kLower == "connection" || kLower == "access-control-allow-origin" {
				continue
			}
			for _, v := range vv {
				if kLower == "content-type" && strings.Contains(v, "force-download") {
					w.Header().Set("Content-Type", "video/mp4")
				} else {
					w.Header().Add(k, v)
				}
			}
		}

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		w.WriteHeader(resp.StatusCode)
		
		io.Copy(w, resp.Body)
	})

	subtitleProxyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		targetUrl := r.URL.Query().Get("url")
		if targetUrl == "" {
			http.Error(w, "url parameter is required", http.StatusBadRequest)
			return
		}

		req, err := http.NewRequest("GET", targetUrl, nil)
		if err != nil {
			http.Error(w, `{"error":"Proxy fetch failed"}`, http.StatusInternalServerError)
			return
		}

		req.Header.Set("User-Agent", "Mozilla/5.0")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			http.Error(w, `{"error":"Proxy fetch failed"}`, http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, "Failed to read subtitle body", http.StatusInternalServerError)
			return
		}

		srtContent := string(bodyBytes)
		
		// Very basic SRT to VTT conversion
		vttContent := "WEBVTT\n\n" + srtContent
		// Replace commas with dots in timestamps: 00:02:17,440 --> 00:02:17.440
		re := regexp.MustCompile(`(\d{2}:\d{2}:\d{2}),(\d{3})`)
		vttContent = re.ReplaceAllString(vttContent, "$1.$2")

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "text/vtt")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(vttContent))
	})

	mux.Handle("/api/proxy/stream", proxyHandler)
	mux.Handle("/api/proxy/stream.mp4", proxyHandler)
	mux.Handle("/api/proxy/stream.m3u8", proxyHandler)
	mux.Handle("/api/proxy/subtitle", subtitleProxyHandler)
}
