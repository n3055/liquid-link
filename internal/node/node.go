package node

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/n3055/liquid-link/internal/api"
	"github.com/n3055/liquid-link/internal/blockchain"
	"github.com/n3055/liquid-link/internal/config"
	"github.com/n3055/liquid-link/internal/consensus"
	"github.com/n3055/liquid-link/internal/ledger"
	"github.com/n3055/liquid-link/internal/pcn"
	"github.com/n3055/liquid-link/internal/peer"
	"github.com/n3055/liquid-link/internal/receipt"
	"github.com/n3055/liquid-link/internal/storage"
	"github.com/n3055/liquid-link/internal/transport"
	"github.com/n3055/liquid-link/internal/wallet"
)

// Node represents the primary execution block for the LiquidLink daemon.
type Node struct {
	cfg          *config.Config
	logger       *slog.Logger
	store        storage.Store
	ledger       ledger.Ledger
	blockchain   blockchain.Blockchain
	consensus    consensus.Engine
	pcnManager   pcn.Manager
	bridge       receipt.Bridge
	transports   []transport.Transport
	peerRegistry peer.Registry
	wallet       wallet.Wallet
	apiServer    *api.Server
	httpServer   *http.Server
}

// NewNode instantiates the node, injecting all concrete placeholders.
func NewNode(cfg *config.Config, logger *slog.Logger) (*Node, error) {
	// 1. Initialize Storage
	store, err := storage.NewBadgerStore(&cfg.Storage)
	if err != nil {
		return nil, fmt.Errorf("failed to load storage: %w", err)
	}

	// TODO: Initialize concrete implementations of components
	var led ledger.Ledger
	var bc blockchain.Blockchain
	var cons consensus.Engine
	var pcnMgr pcn.Manager
	var br receipt.Bridge
	var w wallet.Wallet
	var reg peer.Registry

	transports := make([]transport.Transport, 0)

	apiSrv := api.NewServer(&cfg.Node, led, pcnMgr, br, w)

	httpSrv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Node.Port),
		Handler: apiSrv.Handler(),
	}

	return &Node{
		cfg:          cfg,
		logger:       logger,
		store:        store,
		ledger:       led,
		blockchain:   bc,
		consensus:    cons,
		pcnManager:   pcnMgr,
		bridge:       br,
		transports:   transports,
		peerRegistry: reg,
		wallet:       w,
		apiServer:    apiSrv,
		httpServer:   httpSrv,
	}, nil
}

// Start boots all transports, network listener ports, background synchronization, and the HTTP server.
func (n *Node) Start(ctx context.Context) error {
	n.logger.Info("starting LiquidLink node", slog.String("node_id", n.cfg.Node.ID))

	// Start all transports
	for _, t := range n.transports {
		if err := t.Start(ctx); err != nil {
			return fmt.Errorf("failed to start transport %s: %w", t.Medium(), err)
		}
	}

	// Start API Server
	go func() {
		n.logger.Info("starting HTTP API Server", slog.Int("port", n.cfg.Node.Port))
		if err := n.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			n.logger.Error("HTTP server failure", slog.Any("error", err))
		}
	}()

	return nil
}

// Stop executes graceful termination of all active transports, DB hooks, and HTTP servers.
func (n *Node) Stop(ctx context.Context) error {
	n.logger.Info("stopping LiquidLink node")

	// Shutdown HTTP Server
	if err := n.httpServer.Shutdown(ctx); err != nil {
		n.logger.Error("HTTP server shutdown error", slog.Any("error", err))
	}

	// Stop all transports
	for _, t := range n.transports {
		if err := t.Stop(ctx); err != nil {
			n.logger.Error("failed to stop transport", slog.Any("medium", t.Medium()), slog.Any("error", err))
		}
	}

	// Close database
	if err := n.store.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}

	return nil
}
