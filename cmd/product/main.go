// Package main is the entry point for the product service.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/vukieuhaihoa/my-go-coffeeshop/cmd/product/config"
)

func main() {
	defaultConfigPath := "config.yml"
	if p := os.Getenv("CONFIG_PATH"); p != "" {
		defaultConfigPath = p
	}

	configPath := flag.String("config", defaultConfigPath, "Path to the configuration file")
	flag.Parse()

	cfg, err := config.NewConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to create config: %v", err)
	}

	fmt.Printf("Hello from %s version %s\n", cfg.Name, cfg.Version)
	fmt.Printf("This service is running on %s:%s\n", cfg.Host, cfg.Port)
	fmt.Printf("Log level is set to %s\n", cfg.Level)
}
