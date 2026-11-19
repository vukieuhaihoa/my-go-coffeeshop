// Package app provides the main application logic for the product service.
package app

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"time"

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

// Run starts the product service application with the provided context and configuration.
// It initializes the logger and logs the service start information.
// Parameters:
//   - ctx: The context for managing the application's lifecycle.
//   - cfg: The configuration for the product service.
//
// Returns:
//   - An error if the gRPC server fails to start or encounters an issue during execution.
func Run(ctx context.Context, cfg *config.Config) error {
	log, err := logger.NewSlogLogger(os.Stdout, cfg.Level)
	if err != nil {
		return err
	}

	log.Info("service starting",
		"app_name", cfg.Name,
		"app_version", cfg.Version,
		"grpc_host", cfg.Host,
		"grpc_port", cfg.Port,
	)

	addr := net.JoinHostPort(cfg.Host, cfg.Port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", addr, err)
	}

	srv := grpc.NewServer(grpc.ChainUnaryInterceptor(logUnary(log), recoverUnary(log)))
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

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- srv.Serve(lis)
	}()
	log.Info("gRPC server is listening", "address", lis.Addr().String())

	var serveFailure error

	select {
	case err := <-serveErr:
		serveFailure = fmt.Errorf("grpc serve: %w", err)
		log.Error("gRPC server stopped with error", "error", err)
	case <-ctx.Done():
		log.Info("shutting down", "reason", context.Cause(ctx))
	}

	// ③ Báo "tôi sắp nghỉ" TRƯỚC khi bắt đầu tắt
	hs.Shutdown()

	if err := gracefulStop(srv, shutdownTimeout); err != nil {
		return errors.Join(serveFailure, err)
	}

	if serveFailure != nil {
		return serveFailure
	}

	log.Info("bye")
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
