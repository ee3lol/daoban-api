package main

import (
	"ee3lol/daoban-api/internal/config"
	"ee3lol/daoban-api/internal/server"
)

func main() {
	cfg := config.Load()
	srv := server.New(cfg)
	srv.Start()
}
