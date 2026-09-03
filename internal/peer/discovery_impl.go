package peer

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// httpPingDiscovery implements Discovery by periodically pinging registered remote peer HTTP endpoints.
type httpPingDiscovery struct {
	mu       sync.RWMutex
	reg      Registry
	selfID   string
	client   *http.Client
	stopChan chan struct{}
	running  bool
}

// [Peer Discovery] NewHTTPPingDiscovery creates a new HTTP-based peer discovery and heartbeat service.
func NewHTTPPingDiscovery(reg Registry, selfID string) Discovery {
	return &httpPingDiscovery{
		reg:    reg,
		selfID: selfID,
		client: &http.Client{
			Timeout: 3 * time.Second,
		},
		stopChan: make(chan struct{}),
	}
}

// [Peer Discovery] Start launches the background peer ping and heartbeat monitoring loop.
func (d *httpPingDiscovery) Start(ctx context.Context) error {
	d.mu.Lock()
	if d.running {
		d.mu.Unlock()
		return fmt.Errorf("peer discovery service already running")
	}
	d.running = true
	d.mu.Unlock()

	go d.pingLoop(ctx)
	return nil
}

// [Peer Discovery] Stop terminates the peer discovery heartbeat loop.
func (d *httpPingDiscovery) Stop(ctx context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.running {
		return nil
	}

	d.running = false
	close(d.stopChan)
	return nil
}

// pingLoop periodically queries registered remote peers and updates their connection status.
func (d *httpPingDiscovery) pingLoop(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-d.stopChan:
			return
		case <-ticker.C:
			d.pingOnce(ctx)
		}
	}
}

// [Peer Discovery] pingOnce performs a single round of ping requests against all registered remote peers.
func (d *httpPingDiscovery) pingOnce(ctx context.Context) {
	peers, err := d.reg.ListPeers(ctx)
	if err != nil {
		return
	}

	for _, p := range peers {
		if p.ID == d.selfID || p.Endpoint == "" {
			continue
		}

		pingURL := fmt.Sprintf("%s/api/v1/status", p.Endpoint)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, pingURL, nil)
		if err != nil {
			continue
		}

		resp, err := d.client.Do(req)
		connected := (err == nil && resp.StatusCode == http.StatusOK)
		if resp != nil {
			_ = resp.Body.Close()
		}

		p.Connected = connected
		p.LastSeen = time.Now()
		_ = d.reg.RegisterPeer(ctx, p)
	}
}
