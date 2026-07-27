// Package main implements the node daemon entry point.
package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/n3055/liquid-link/internal/config"
	"github.com/n3055/liquid-link/internal/node"
)

func main() {
	configPath := flag.String("config", "configs/config.yaml", "path to node configuration file")
	flag.Parse()

	// Initialize structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	logger.Info("initializing LiquidLink daemon")

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		logger.Error("failed to load configuration", slog.Any("error", err))
		os.Exit(1)
	}

	// Create root execution context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize Node
	n, err := node.NewNode(cfg, logger)
	if err != nil {
		logger.Error("failed to initialize node components", slog.Any("error", err))
		os.Exit(1)
	}

	// Listen for OS interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start Node
	if err := n.Start(ctx); err != nil {
		logger.Error("node failed to start", slog.Any("error", err))
		os.Exit(1)
	}

	// Wait for terminate signal
	sig := <-sigChan
	logger.Info("shutdown signal received", slog.String("signal", sig.String()))

	// Graceful shutdown context
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := n.Stop(shutdownCtx); err != nil {
		logger.Error("failed to stop node cleanly", slog.Any("error", err))
		os.Exit(1)
	}

	logger.Info("node exited successfully")
}
