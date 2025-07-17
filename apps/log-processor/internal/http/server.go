package http

import (
	"context"
	"log-processor/internal/interfaces"
	"log/slog"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Server struct {
	configService interfaces.ConfigService
	server        *http.Server
	log           *slog.Logger
}

func NewServer(configService interfaces.ConfigService, log *slog.Logger) *Server {
	mux := http.NewServeMux()
	
	// Add Prometheus metrics endpoint
	mux.Handle("/metrics", promhttp.Handler())
	
	// Add health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	server := &http.Server{
		Addr:    configService.GetConfig().HTTP.Addr,
		Handler: mux,
	}

	return &Server{
		configService: configService,
		server:        server,
		log:           log,
	}
}

func (s *Server) Start(ctx context.Context) error {
	addr := s.configService.GetConfig().HTTP.Addr
	s.log.Info("Starting HTTP server", "addr", addr)

	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	s.log.Info("Stopping HTTP server")
	return s.server.Shutdown(ctx)
} 