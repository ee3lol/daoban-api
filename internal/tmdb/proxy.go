package tmdb

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"ee3lol/daoban-api/internal/config"
)

const TMDB_BASE_URL = "https://api.themoviedb.org/3"

func NewProxy(cfg *config.Config) *httputil.ReverseProxy {
	targetURL, err := url.Parse(TMDB_BASE_URL)
	if err != nil {
		log.Fatalf("Failed to parse TMDB_BASE_URL: %v", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		
		req.Header.Set("Authorization", "Bearer "+cfg.TMDBToken)
		req.Header.Set("Accept", "application/json")
		
		req.Host = targetURL.Host
	}

	return proxy
}
