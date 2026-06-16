package main

import (
	"encoding/json"
	"os"
	"testing"
)

func TestParseConfig(t *testing.T) {
	jsonData := `{
		"global": {
			"connect_timeout_seconds": 10,
			"execute_timeout_seconds": 120,
			"retry_count": 2,
			"retry_delay_seconds": 5,
			"log_dir": "/var/log/ups",
			"log_max_days": 30
		},
		"default_script": [
			"sync",
			"shutdown -h now"
		],
		"servers": [
			{
				"host": "192.168.1.10",
				"port": 22,
				"user": "root",
				"password": "secret1",
				"script": [
					"systemctl stop nginx",
					"shutdown -h now"
				]
			},
			{
				"host": "192.168.1.11",
				"user": "admin",
				"password": "secret2"
			}
		]
	}`

	cfg, err := ParseConfig([]byte(jsonData))
	if err != nil {
		t.Fatalf("ParseConfig failed: %v", err)
	}

	// FillDefaults must be called before checking default values
	cfg.FillDefaults()

	// Check global defaults
	if cfg.Global.SSHPort != 22 {
		t.Errorf("expected default SSHPort=22, got %d", cfg.Global.SSHPort)
	}
	if cfg.Global.ConnectTimeoutSeconds != 10 {
		t.Errorf("expected ConnectTimeoutSeconds=10, got %d", cfg.Global.ConnectTimeoutSeconds)
	}

	// Check server 1
	if len(cfg.Servers) != 2 {
		t.Fatalf("expected 2 servers, got %d", len(cfg.Servers))
	}
	s1 := cfg.Servers[0]
	if s1.Host != "192.168.1.10" {
		t.Errorf("expected host 192.168.1.10, got %s", s1.Host)
	}
	if s1.Port != 22 {
		t.Errorf("expected port 22, got %d", s1.Port)
	}
	if len(s1.Script) != 2 {
		t.Errorf("expected 2 script commands, got %d", len(s1.Script))
	}

	// Check server 2 — should inherit default port and default script
	s2 := cfg.Servers[1]
	if s2.Host != "192.168.1.11" {
		t.Errorf("expected host 192.168.1.11, got %s", s2.Host)
	}
	if s2.Port != 22 {
		t.Errorf("expected default port 22, got %d", s2.Port)
	}
	if len(s2.Script) != 2 {
		t.Errorf("expected 2 default script commands, got %d", len(s2.Script))
	}
	if s2.Script[0] != "sync" {
		t.Errorf("expected 'sync', got '%s'", s2.Script[0])
	}
}

func TestParseConfigFile(t *testing.T) {
	// Write a temp file
	tmpFile := "test_config.json"
	jsonData := `{
		"global": {
			"connect_timeout_seconds": 5
		},
		"default_script": ["shutdown -h now"],
		"servers": [
			{"host": "10.0.0.1", "user": "root", "password": "pwd"}
		]
	}`
	if err := os.WriteFile(tmpFile, []byte(jsonData), 0644); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}
	defer os.Remove(tmpFile)

	cfg, err := ParseConfigFile(tmpFile)
	if err != nil {
		t.Fatalf("ParseConfigFile failed: %v", err)
	}
	cfg.FillDefaults()

	if len(cfg.Servers) != 1 {
		t.Fatalf("expected 1 server, got %d", len(cfg.Servers))
	}
	s := cfg.Servers[0]
	if s.Port != 22 {
		t.Errorf("expected default port 22, got %d", s.Port)
	}
	if s.ConnectTimeoutSeconds != 5 {
		t.Errorf("expected ConnectTimeoutSeconds=5, got %d", s.ConnectTimeoutSeconds)
	}
}

func TestParseConfigInvalidJSON(t *testing.T) {
	_, err := ParseConfig([]byte("not json"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestParseConfigMissingServers(t *testing.T) {
	jsonData := `{"global": {}, "default_script": [], "servers": []}`
	cfg, err := ParseConfig([]byte(jsonData))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cfg.FillDefaults()
	// Empty servers is valid — nothing to do
	if len(cfg.Servers) != 0 {
		t.Errorf("expected 0 servers, got %d", len(cfg.Servers))
	}
}

func TestConfigMarshalRoundTrip(t *testing.T) {
	orig := &Config{
		Global: GlobalConfig{
			SSHPort:               22,
			ConnectTimeoutSeconds: 10,
			ExecuteTimeoutSeconds: 60,
			RetryCount:            2,
			RetryDelaySeconds:     5,
			LogDir:                ".",
			LogMaxDays:            30,
		},
		DefaultScript: []string{"sync", "shutdown -h now"},
		Servers: []ServerConfig{
			{Host: "1.2.3.4", User: "root", Password: "pw", Port: 22},
		},
	}
	orig.FillDefaults()

	data, err := json.MarshalIndent(orig, "", "  ")
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	parsed, err := ParseConfig(data)
	if err != nil {
		t.Fatalf("parse back failed: %v", err)
	}
	parsed.FillDefaults()

	if len(parsed.Servers) != 1 {
		t.Fatalf("round-trip server count mismatch")
	}
	if parsed.Servers[0].Host != "1.2.3.4" {
		t.Errorf("round-trip host mismatch")
	}
}
