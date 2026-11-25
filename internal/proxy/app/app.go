// Package app provides the main application logic for the proxy service.
package app

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/vukieuhaihoa/my-go-coffeeshop/cmd/proxy/config"
	"github.com/vukieuhaihoa/my-go-coffeeshop/internal/proxy/gateway"
	"github.com/vukieuhaihoa/my-go-coffeeshop/pkg/logger"
)

const (
	shutdownTimeout = 5 * time.Second
	requestTimeout  = 3 * time.Second
)

// App represents the main application for the proxy service.
type App struct {
	cfg *config.Config
	log logger.Logger
}

// New creates a new instance of the App with the provided configuration and logger.
// Parameters:
//   - cfg: The configuration for the proxy service.
//   - log: The logger to be used for logging within the application.
//
// Returns:
//   - A pointer to an App instance.
func New(cfg *config.Config, log logger.Logger) *App {
	return &App{
		cfg: cfg,
		log: log,
	}
}

// Run starts the HTTP server and handles graceful shutdown on receiving termination signals.
// Parameters:
//   - ctx: The context for managing the application's lifecycle.
//   - lis: The network listener for the HTTP server.
//
// Returns:
//   - An error if there was an issue running the application.
func (app *App) Run(ctx context.Context, lis net.Listener) error {
	conn, err := grpc.NewClient(app.cfg.ProductAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("create product client: %w", err)
	}
	defer func() {
		if cerr := conn.Close(); cerr != nil {
			app.log.Error("failed to close gRPC connection", "error", cerr)
		}
	}()

	gw, err := gateway.New(ctx, conn)
	if err != nil {
		return fmt.Errorf("create gateway: %w", err)
	}

	handler := gateway.LogRequests(
		app.log,
		gateway.AllowCORS(
			app.cfg.AllowedOrigins,
			gateway.WithTimeout(
				requestTimeout,
				gw,
			),
		),
	)

	srv := &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      shutdownTimeout,
		IdleTimeout:       120 * time.Second,
		// MaxHeaderBytes:   1 << 20, // 1 MB
	}

	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		app.log.Info("starting HTTP server", "address", lis.Addr().String(), "app", app.cfg.Name, "version", app.cfg.Version)
		if err := srv.Serve(lis); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("HTTP server listen: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		<-gctx.Done()
		app.log.Info("shutting down", "reason", context.Cause(gctx))
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			return fmt.Errorf("HTTP server shutdown: %w", err)
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return fmt.Errorf("run error: %w", err)
	}

	app.log.Info("bye")
	return nil
}
