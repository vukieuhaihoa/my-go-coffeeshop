// Package main is the entry point for the proxy service.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/vukieuhaihoa/my-go-coffeeshop/cmd/proxy/config"
	"github.com/vukieuhaihoa/my-go-coffeeshop/internal/proxy/app"
	"github.com/vukieuhaihoa/my-go-coffeeshop/pkg/logger"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}

func run() error {
	defaultConfigPath := "config.yml"
	if p := os.Getenv("CONFIG_PATH"); p != "" {
		defaultConfigPath = p
	}

	configPath := flag.String("config", defaultConfigPath, "Path to the configuration file")
	flag.Parse()

	cfg, err := config.NewConfig(*configPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	lg, err := logger.NewSlogLogger(os.Stdout, cfg.Level)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	lis, err := net.Listen("tcp", net.JoinHostPort(cfg.Host, cfg.Port))
	if err != nil {
		return fmt.Errorf("failed to listen on %s:%s: %w", cfg.Host, cfg.Port, err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		stop()
	}()

	return app.New(cfg, lg).Run(ctx, lis)
}
