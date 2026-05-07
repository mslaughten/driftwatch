package healthcheck

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// Server wraps an HTTP server that serves the health endpoint.
type Server struct {
	handler *Handler
	httpSrv *http.Server
}

// NewServer creates a Server listening on the given address (e.g. ":9090").
func NewServer(addr string, h *Handler) *Server {
	mux := http.NewServeMux()
	mux.Handle("/healthz", h)
	return &Server{
		handler: h,
		httpSrv: &http.Server{
			Addr:         addr,
			Handler:      mux,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
		},
	}
}

// Start begins serving in a background goroutine.
// It returns an error if the server fails to bind.
func (s *Server) Start() error {
	errCh := make(chan error, 1)
	go func() {
		if err := s.httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("healthcheck server: %w", err)
		}
	}()
	// Give the server a moment to fail fast on bind errors.
	select {
	case err := <-errCh:
		return err
	case <-time.After(50 * time.Millisecond):
		return nil
	}
}

// Stop gracefully shuts down the server.
func (s *Server) Stop(ctx context.Context) error {
	return s.httpSrv.Shutdown(ctx)
}
