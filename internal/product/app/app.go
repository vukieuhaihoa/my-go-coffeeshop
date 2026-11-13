// Package app provides the main application logic for the product service.
package app

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/vukieuhaihoa/my-go-coffeeshop/cmd/product/config"
	"github.com/vukieuhaihoa/my-go-coffeeshop/internal/product/httpapi"
	"github.com/vukieuhaihoa/my-go-coffeeshop/pkg/logger"
)

const (
	// shutdownTimeout defines the maximum duration to wait for the application to shut down gracefully.
	shutdownTimeout = 5 * time.Second
)

// Run starts the product service application with the provided context and configuration.
// It initializes the logger, sets up the HTTP server, and handles graceful shutdown on context cancellation.
// Parameters:
//   - ctx: The context for managing the application's lifecycle.
//   - cfg: The configuration for the product service.
//
// Returns:
//   - An error if the application fails to start or encounters an issue during execution.
func Run(ctx context.Context, cfg *config.Config) error {
	log, err := logger.NewSlogLogger(os.Stdout, cfg.Level)
	if err != nil {
		return err
	}

	log.Info("service started",
		"app_name", cfg.Name,
		"app_version", cfg.Version,
		"http_host", cfg.Host,
		"http_port", cfg.Port,
	)

	h := httpapi.NewHandler(log)

	srv := &http.Server{
		Addr:              net.JoinHostPort(cfg.Host, cfg.Port),
		Handler:           logRequests(log, h.Routes()),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      shutdownTimeout,
		IdleTimeout:       120 * time.Second,
	}
	ln, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		return err
	}

	serveErr := make(chan error, 1)
	go func() {
		log.Info("HTTP server is listening", "address", srv.Addr)
		if err := srv.Serve(ln); !errors.Is(err, http.ErrServerClosed) {
			serveErr <- err
		}
		close(serveErr)
	}()

	select {
	case err, ok := <-serveErr:
		if !ok {
			return errors.New("server stopped unexpectedly")
		}
		return fmt.Errorf("http server: %w", err)
	case <-ctx.Done():
		log.Info("shutting down", "reason", context.Cause(ctx))
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		_ = srv.Close() // Force close if shutdown fails
		return fmt.Errorf("shutdown HTTP server: %w", err)
	}

	log.Info("bye")
	return nil
}
