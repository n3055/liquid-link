package peer

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/n3055/liquid-link/internal/config"
	"github.com/n3055/liquid-link/internal/storage"
)

// [Peer Discovery] TestPeerDiscoveryAndSelfExclusion verifies registering peers and HTTP discovery pinging.
func TestPeerDiscoveryAndSelfExclusion(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "peer-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	store, err := storage.NewBadgerStore(&config.StorageConfig{
		Path:             filepath.Join(tempDir, "badger"),
		ValueLogFileSize: 10485760,
	})
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	reg := NewRegistry(store)

	// Register self and a mock remote peer
	selfInfo := &Info{ID: "self-node", Endpoint: "http://localhost:8080", Connected: true}
	if err := reg.RegisterPeer(ctx, selfInfo); err != nil {
		t.Fatalf("failed to register self: %v", err)
	}

	// Mock peer server returning 200 OK for /api/v1/status
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/status" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"ok"}`))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer mockServer.Close()

	remoteInfo := &Info{ID: "remote-node-1", Endpoint: mockServer.URL, Connected: false}
	if err := reg.RegisterPeer(ctx, remoteInfo); err != nil {
		t.Fatalf("failed to register remote peer: %v", err)
	}

	disc := NewHTTPPingDiscovery(reg, "self-node")
	if err := disc.Start(ctx); err != nil {
		t.Fatalf("failed to start discovery: %v", err)
	}

	discImpl := disc.(*httpPingDiscovery)
	discImpl.pingOnce(ctx) // Run one manual ping pass

	if err := disc.Stop(ctx); err != nil {
		t.Fatalf("failed to stop discovery: %v", err)
	}

	updatedRemote, err := reg.GetPeer(ctx, "remote-node-1")
	if err != nil {
		t.Fatalf("failed to get remote peer: %v", err)
	}
	if !updatedRemote.Connected {
		t.Errorf("expected remote peer to be marked Connected=true after ping")
	}
}
