// Package main is the entry point for the product service.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/vukieuhaihoa/my-go-coffeeshop/cmd/product/config"
	"github.com/vukieuhaihoa/my-go-coffeeshop/internal/product/app"
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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		stop()
	}()

	return app.Run(ctx, cfg)
}
