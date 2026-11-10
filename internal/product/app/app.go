// Package app provides the main application logic for the product service.
package app

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/vukieuhaihoa/my-go-coffeeshop/cmd/product/config"
	"github.com/vukieuhaihoa/my-go-coffeeshop/pkg/logger"
)

const (
	// shutdownTimeout defines the maximum duration to wait for the application to shut down gracefully.
	shutdownTimeout = 5 * time.Second
)

// Run starts the application with the provided context and configuration.
// It initializes the logger, starts the heartbeat goroutine, and waits for a shutdown signal.
// Parameters:
//   - ctx: The context to control the application's lifecycle.
//   - cfg: The configuration for the application.
//
// Returns:
//   - An error if the application fails to run or shuts down with an error.
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

	var wg sync.WaitGroup
	wg.Go(func() {
		heartbeat(ctx, log, 2*time.Second)
	})

	<-ctx.Done()
	log.Info("shutting down", "reason", context.Cause(ctx))

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	select {
	case <-done:
	case <-shutdownCtx.Done():
		log.Warn("shutdown timeout exceeded, forcing exit")
		return fmt.Errorf("shutdown: %w", shutdownCtx.Err())
	}

	log.Info("bye")
	return nil
}

func heartbeat(ctx context.Context, log logger.Logger, every time.Duration) {
	ticker := time.NewTicker(every)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info("heartbeat stopped")
			return
		case t := <-ticker.C:
			log.Debug("heartbeat", "at", t.Format(time.TimeOnly))
		}
	}
}
