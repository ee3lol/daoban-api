package server

import (
	"net/http"

	"ee3lol/daoban-api/internal/config"
	"ee3lol/daoban-api/internal/tmdb"
)

func setupRoutes(mux *http.ServeMux, cfg *config.Config) {
	tmdbProxy := tmdb.NewProxy(cfg)
	mux.Handle("/", tmdbProxy)

	setupProxyRoutes(mux)
	setupAPIRoutes(mux, cfg)
}
