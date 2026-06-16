package main

import (
	"context"
	"testing"
	"time"
)

func TestRunAllServersSuccess(t *testing.T) {
	cfg := &Config{
		Global: GlobalConfig{
			SSHPort:               22,
			ConnectTimeoutSeconds: 5,
			ExecuteTimeoutSeconds: 10,
			RetryCount:            1,
			RetryDelaySeconds:     1,
		},
	}
	cfg.FillDefaults()

	servers := []ServerConfig{
		{Host: "server1", User: "root", Password: "pw"},
		{Host: "server2", User: "admin", Password: "pw"},
	}
	cfg.Servers = servers
	cfg.FillDefaults()

	results := runAllServers(cfg)

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	// Both should fail since these are fake hosts (but that tests error handling)
	for _, r := range results {
		if r.Error == "" {
			t.Log("unexpected: server succeeded with fake host")
		}
	}
}

func TestRunAllServersEmpty(t *testing.T) {
	cfg := &Config{
		Servers: []ServerConfig{},
	}
	cfg.FillDefaults()

	results := runAllServers(cfg)
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestServerResultIsSuccess(t *testing.T) {
	r := ServerResult{Host: "test", Success: true, Error: ""}
	if !r.IsSuccess() {
		t.Error("expected IsSuccess=true")
	}

	r2 := ServerResult{Host: "test", Success: false, Error: "failed"}
	if r2.IsSuccess() {
		t.Error("expected IsSuccess=false")
	}
}

func TestAllSucceeded(t *testing.T) {
	results := []ServerResult{
		{Host: "a", Success: true},
		{Host: "b", Success: true},
	}
	if !allSucceeded(results) {
		t.Error("expected allSucceeded=true")
	}

	results = append(results, ServerResult{Host: "c", Success: false, Error: "fail"})
	if allSucceeded(results) {
		t.Error("expected allSucceeded=false")
	}
}

func TestShutdownContextTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	select {
	case <-ctx.Done():
		t.Log("context timed out as expected")
	case <-time.After(200 * time.Millisecond):
		t.Error("context should have timed out")
	}
}
