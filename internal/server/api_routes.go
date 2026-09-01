package server

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"ee3lol/daoban-api/internal/config"
	"ee3lol/daoban-api/internal/scraper"
	"ee3lol/daoban-api/internal/scraper/providers/oneembed"
)

func setupAPIRoutes(mux *http.ServeMux, cfg *config.Config) {
	scraperImpl := oneembed.NewScraper(cfg.APIBaseURL)
	authMiddleware := APIKeyAuth(cfg.APIKey)

	// Custom router for /api/ routes
	apiMux := http.NewServeMux()

	apiMux.HandleFunc("/api/movie/", func(w http.ResponseWriter, r *http.Request) {
		tmdbId := strings.TrimPrefix(r.URL.Path, "/api/movie/")
		if tmdbId == "" {
			http.Error(w, "tmdbId required", http.StatusBadRequest)
			return
		}

		details, err := scraperImpl.GetDetails(tmdbId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		sources, err := scraperImpl.GetStreamSources("movie", tmdbId, 1, 1)
		if err != nil {
			sources = []scraper.StreamSource{}
		}

		resp := map[string]interface{}{
			"success": true,
			"media":   details,
			"sources": sources,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	apiMux.HandleFunc("/api/tv/", func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/tv/"), "/")
		if len(parts) != 3 {
			http.Error(w, "invalid path format, expected /api/tv/{tmdbId}/{season}/{episode}", http.StatusBadRequest)
			return
		}

		tmdbId := parts[0]
		season, _ := strconv.Atoi(parts[1])
		episode, _ := strconv.Atoi(parts[2])

		details, err := scraperImpl.GetDetails(tmdbId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		sources, err := scraperImpl.GetStreamSources("tv", tmdbId, season, episode)
		if err != nil {
			sources = []scraper.StreamSource{}
		}

		resp := map[string]interface{}{
			"success": true,
			"media":   details,
			"season":  season,
			"episode": episode,
			"sources": sources,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	apiMux.HandleFunc("/api/subtitles/", func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/subtitles/"), "/")
		if len(parts) != 2 {
			http.Error(w, "invalid path format, expected /api/subtitles/{type}/{tmdbId}", http.StatusBadRequest)
			return
		}

		mediaType := parts[0]
		tmdbId := parts[1]

		season := 1
		episode := 1
		if sStr := r.URL.Query().Get("s"); sStr != "" {
			season, _ = strconv.Atoi(sStr)
		}
		if eStr := r.URL.Query().Get("e"); eStr != "" {
			episode, _ = strconv.Atoi(eStr)
		}

		subtitles, err := scraperImpl.GetSubtitles(mediaType, tmdbId, season, episode)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if subtitles == nil {
			subtitles = []scraper.Subtitle{}
		}

		resp := map[string]interface{}{
			"success":   true,
			"subtitles": subtitles,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// Wrap apiMux with authMiddleware and register on main mux
	mux.Handle("/api/", authMiddleware(apiMux))
}
