package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the top-level configuration structure for a LiquidLink node.
type Config struct {
	Node       NodeConfig      `yaml:"node"`
	Storage    StorageConfig   `yaml:"storage"`
	Consensus  ConsensusConfig `yaml:"consensus"`
	Transports TransportConfig `yaml:"transports"`
	PCN        PCNConfig       `yaml:"pcn"`
}

// NodeConfig contains identity and logging settings for the node.
type NodeConfig struct {
	ID       string `yaml:"id"`
	Address  string `yaml:"address"`
	Port     int    `yaml:"port"`
	LogLevel string `yaml:"log_level"`
}

// StorageConfig specifies configurations for local persistence.
type StorageConfig struct {
	Path             string `yaml:"path"`
	ValueLogFileSize int64  `yaml:"value_log_file_size"`
}

// ConsensusConfig defines consensus timing and voting parameters.
type ConsensusConfig struct {
	LeaderRotationInterval string  `yaml:"leader_rotation_interval"`
	VoteTimeout            string  `yaml:"vote_timeout"`
	MajorityThreshold      float64 `yaml:"majority_threshold"`
}

// TransportConfig groups settings for all supported communication mediums.
type TransportConfig struct {
	Internet  InternetConfig  `yaml:"internet"`
	Bluetooth BluetoothConfig `yaml:"bluetooth"`
}

// InternetConfig configures the TCP/IP networking layer.
type InternetConfig struct {
	Enabled        bool     `yaml:"enabled"`
	BindAddress    string   `yaml:"bind_address"`
	BootstrapPeers []string `yaml:"bootstrap_peers"`
}

// BluetoothConfig configures the Bluetooth Mesh networking layer.
type BluetoothConfig struct {
	Enabled     bool   `yaml:"enabled"`
	DeviceID    string `yaml:"device_id"`
	ServiceUUID string `yaml:"service_uuid"`
}

// PCNConfig configures payment channel network options.
type PCNConfig struct {
	ChannelReserve    float64 `yaml:"channel_reserve"`
	RoutingFeeFlat    int64   `yaml:"routing_fee_flat"`
	RoutingFeePercent float64 `yaml:"routing_fee_percent"`
}

// Load reads and parses configuration from a YAML file.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
