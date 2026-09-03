package node

import (
	"context"
	"io/ioutil"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/n3055/liquid-link/internal/config"
)

// [Consensus Block] TestNodeStartStopAndBlockProducer verifies node boot, block production ticker, and shutdown.
func TestNodeStartStopAndBlockProducer(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "liquidlink-node-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfg := &config.Config{
		Node: config.NodeConfig{
			ID:   "node-test",
			Port: 18080,
		},
		Storage: config.StorageConfig{
			Path:             filepath.Join(tempDir, "badger"),
			ValueLogFileSize: 10485760,
		},
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	n, err := NewNode(cfg, logger)
	if err != nil {
		t.Fatalf("failed to create node: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := n.Start(ctx); err != nil {
		t.Fatalf("failed to start node: %v", err)
	}

	// Allow block producer to run briefly
	time.Sleep(100 * time.Millisecond)

	// Check block height is at least genesis (>= 1)
	h, err := n.blockchain.Height(ctx)
	if err != nil {
		t.Fatalf("failed to get block height: %v", err)
	}
	if h < 1 {
		t.Errorf("expected block height >= 1, got %d", h)
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := n.Stop(shutdownCtx); err != nil {
		t.Fatalf("failed to stop node: %v", err)
	}
}
