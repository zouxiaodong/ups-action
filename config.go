package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// GlobalConfig holds global settings.
type GlobalConfig struct {
	SSHPort               int    `json:"ssh_port"`
	ConnectTimeoutSeconds int    `json:"connect_timeout_seconds"`
	ExecuteTimeoutSeconds int    `json:"execute_timeout_seconds"`
	RetryCount            int    `json:"retry_count"`
	RetryDelaySeconds     int    `json:"retry_delay_seconds"`
	LogDir                string `json:"log_dir"`
	LogMaxDays            int    `json:"log_max_days"`
}

// ServerConfig holds per-server connection and script settings.
type ServerConfig struct {
	Host                  string   `json:"host"`
	Port                  int      `json:"port"`
	User                  string   `json:"user"`
	Password              string   `json:"password"`
	Script                []string `json:"script,omitempty"`
	ConnectTimeoutSeconds int      `json:"-"`
	ExecuteTimeoutSeconds int      `json:"-"`
}

// Config is the top-level configuration.
type Config struct {
	Global        GlobalConfig   `json:"global"`
	DefaultScript []string       `json:"default_script"`
	Servers       []ServerConfig `json:"servers"`
}

// FillDefaults fills in default values for missing settings.
func (c *Config) FillDefaults() {
	if c.Global.SSHPort == 0 {
		c.Global.SSHPort = 22
	}
	if c.Global.ConnectTimeoutSeconds == 0 {
		c.Global.ConnectTimeoutSeconds = 10
	}
	if c.Global.ExecuteTimeoutSeconds == 0 {
		c.Global.ExecuteTimeoutSeconds = 120
	}
	if c.Global.RetryDelaySeconds == 0 {
		c.Global.RetryDelaySeconds = 5
	}
	if c.Global.LogDir == "" {
		c.Global.LogDir = "."
	}
	if c.Global.LogMaxDays == 0 {
		c.Global.LogMaxDays = 30
	}

	for i := range c.Servers {
		if c.Servers[i].Port == 0 {
			c.Servers[i].Port = c.Global.SSHPort
		}
		if len(c.Servers[i].Script) == 0 {
			c.Servers[i].Script = make([]string, len(c.DefaultScript))
			copy(c.Servers[i].Script, c.DefaultScript)
		}
		c.Servers[i].ConnectTimeoutSeconds = c.Global.ConnectTimeoutSeconds
		c.Servers[i].ExecuteTimeoutSeconds = c.Global.ExecuteTimeoutSeconds
	}
}

// ParseConfig parses a JSON config from bytes.
func ParseConfig(data []byte) (*Config, error) {
	cfg := &Config{}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	return cfg, nil
}

// ParseConfigFile reads and parses a JSON config file.
func ParseConfigFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
	}
	return ParseConfig(data)
}
