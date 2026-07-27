package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	tempFile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	configContent := `
node:
  id: "test-node"
  address: "LLTESTADDRESS"
  port: 9000
  log_level: "debug"
storage:
  path: "/tmp/badger-test"
  value_log_file_size: 1024
consensus:
  leader_rotation_interval: "10s"
  vote_timeout: "1s"
  majority_threshold: 0.5
transports:
  internet:
    enabled: true
    bind_address: "127.0.0.1:9001"
    bootstrap_peers:
      - "127.0.0.1:9002"
  bluetooth:
    enabled: false
    device_id: "hci0"
    service_uuid: "test-uuid"
pcn:
  channel_reserve: 0.1
  routing_fee_flat: 5
  routing_fee_percent: 0.01
`

	if _, err := tempFile.Write([]byte(configContent)); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tempFile.Close()

	cfg, err := Load(tempFile.Name())
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.Node.ID != "test-node" {
		t.Errorf("expected node ID 'test-node', got %q", cfg.Node.ID)
	}
	if cfg.Node.Port != 9000 {
		t.Errorf("expected node port 9000, got %d", cfg.Node.Port)
	}
	if !cfg.Transports.Internet.Enabled {
		t.Errorf("expected internet transport to be enabled")
	}
}
