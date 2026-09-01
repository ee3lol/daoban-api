package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ee3lol/daoban-api/internal/config"
)

type Server struct {
	httpServer *http.Server
	cfg        *config.Config
}

func New(cfg *config.Config) *Server {
	mux := http.NewServeMux()
	setupRoutes(mux, cfg)

	var handler http.Handler = mux
	handler = CORS(handler)
	handler = Logger(handler)
	handler = Recover(handler)

	return &Server{
		cfg: cfg,
		httpServer: &http.Server{
			Addr:         ":" + cfg.Port,
			Handler:      handler,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}
}

func (s *Server) Start() {
	serverErrors := make(chan error, 1)

	go func() {
		log.Printf("Starting server in %s mode on port %s...", s.cfg.Env, s.cfg.Port)
		serverErrors <- s.httpServer.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		log.Fatalf("Error starting server: %v", err)

	case sig := <-shutdown:
		log.Printf("Start shutdown... Signal: %v", sig)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := s.httpServer.Shutdown(ctx); err != nil {
			log.Printf("Graceful shutdown did not complete in time: %v", err)
			if err := s.httpServer.Close(); err != nil {
				log.Fatalf("Could not stop server gracefully: %v", err)
			}
		}
	}

	log.Println("Server stopped")
}
