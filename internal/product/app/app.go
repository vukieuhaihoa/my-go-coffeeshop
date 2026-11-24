// Package app provides the main application logic for the product service.
package app

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"github.com/vukieuhaihoa/my-go-coffeeshop/cmd/product/config"
	productv1 "github.com/vukieuhaihoa/my-go-coffeeshop/gen/go/coffeeshop/product/v1"
	"github.com/vukieuhaihoa/my-go-coffeeshop/internal/product/catalog"
	"github.com/vukieuhaihoa/my-go-coffeeshop/internal/product/grpcapi"
	"github.com/vukieuhaihoa/my-go-coffeeshop/pkg/logger"
)

const (
	// shutdownTimeout defines the maximum duration to wait for the application to shut down gracefully.
	shutdownTimeout = 5 * time.Second
)

// App represents the main application for the product service.
type App struct {
	cfg *config.Config
	log logger.Logger
}

// New creates a new instance of the App with the provided configuration and logger.
// Parameters:
//   - cfg: The configuration for the product service.
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

// Run starts the gRPC server and handles graceful shutdown on receiving termination signals.
// Parameters:
//   - ctx: The context for managing the application's lifecycle.
//   - lis: The network listener for the gRPC server.
//
// Returns:
//   - An error if there was an issue running the application.
func (app *App) Run(ctx context.Context, lis net.Listener) error {
	srv := grpc.NewServer(grpc.ChainUnaryInterceptor(
		logUnary(app.log),
		recoverUnary(app.log)))
	productv1.RegisterProductServiceServer(srv, grpcapi.NewServer(catalog.New()))

	// ① Health: cho người khác hỏi "còn làm việc được không?"
	hs := health.NewServer()
	healthpb.RegisterHealthServer(srv, hs)
	hs.SetServingStatus(
		productv1.ProductService_ServiceDesc.ServiceName,
		healthpb.HealthCheckResponse_SERVING,
	)

	// ② Reflection: cho người khác hỏi "có những API nào?"
	reflection.Register(srv)

	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		app.log.Info("gRPC server is listening",
			"address", lis.Addr().String(),
			"app", app.cfg.Name,
			"version", app.cfg.Version,
		)
		if err := srv.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			return fmt.Errorf("gRPC server serve: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		<-gctx.Done()
		app.log.Info("shutting down", "reason", context.Cause(gctx))

		// ③ Báo "tôi sắp nghỉ" TRƯỚC khi bắt đầu tắt
		hs.Shutdown()
		return gracefulStop(srv, shutdownTimeout)
	})

	if err := g.Wait(); err != nil {
		return fmt.Errorf("application run: %w", err)
	}

	app.log.Info("bye")
	return nil
}

func gracefulStop(srv *grpc.Server, d time.Duration) error {
	done := make(chan struct{})
	go func() {
		srv.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-time.After(d):
		srv.Stop()
		<-done
		return fmt.Errorf("gRPC server stop time out after %s", d)
	}
}
