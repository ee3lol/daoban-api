package server

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"ee3lol/daoban-api/internal/config"
	"ee3lol/daoban-api/internal/scraper"
	"ee3lol/daoban-api/internal/scraper/providers/anidb"
	"ee3lol/daoban-api/internal/scraper/providers/oneembed"
	"ee3lol/daoban-api/internal/tmdb"
)

func setupAPIRoutes(mux *http.ServeMux, cfg *config.Config) {
	scraperImpl := oneembed.NewScraper(cfg.APIBaseURL)
	animeScraper := anidb.NewScraper(cfg.APIBaseURL)
	authMiddleware := APIKeyAuth(cfg.APIKey)

	// Custom router for /api/ routes
	apiMux := http.NewServeMux()

	apiMux.HandleFunc("/api/movie/", func(w http.ResponseWriter, r *http.Request) {
		tmdbId := strings.TrimPrefix(r.URL.Path, "/api/movie/")
		if tmdbId == "" {
			http.Error(w, "tmdbId required", http.StatusBadRequest)
			return
		}

		mediaType := r.URL.Query().Get("type")

		if mediaType == "anime" {
			details, err := tmdb.FetchDetails(cfg, "movie", tmdbId)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			details.Type = "movie"

			animeSources, err := animeScraper.GetStreamSourcesForAnime(details.Title, 1, 1)
			if err != nil || len(animeSources) == 0 {
				log.Printf("[anidb] Failed to get anime movie sources or empty: %v", err)
				// Fallback to oneembed
				sources, _ := scraperImpl.GetStreamSources("movie", tmdbId, 1, 1)
				animeSources = sources
			}

			resp := map[string]interface{}{
				"success": true,
				"media":   details,
				"sources": animeSources,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}

		details, err := tmdb.FetchDetails(cfg, "movie", tmdbId)
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

		mediaType := r.URL.Query().Get("type")
		
		if mediaType == "anime" {
			// Get TMDB details to pass title to anidb
			details, err := tmdb.FetchDetails(cfg, "tv", tmdbId)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			details.Type = "tv"

			animeSources, err := animeScraper.GetStreamSourcesForAnime(details.Title, season, episode)
			if err != nil || len(animeSources) == 0 {
				log.Printf("[anidb] Failed to get anime sources or empty: %v", err)
				// Fallback to oneembed
				sources, _ := scraperImpl.GetStreamSources("tv", tmdbId, season, episode)
				animeSources = sources
			}

			resp := map[string]interface{}{
				"success": true,
				"media":   details,
				"season":  season,
				"episode": episode,
				"sources": animeSources,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}

		details, err := tmdb.FetchDetails(cfg, "tv", tmdbId)
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
