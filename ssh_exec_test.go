package main

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestBuildScript(t *testing.T) {
	cmds := []string{
		"sync",
		"shutdown -h now",
	}
	script := buildScript(cmds)

	if !strings.Contains(script, "set -e") {
		t.Error("script should contain 'set -e'")
	}
	if !strings.Contains(script, "sync") {
		t.Error("script should contain 'sync'")
	}
	if !strings.Contains(script, "shutdown -h now") {
		t.Error("script should contain 'shutdown -h now'")
	}
	t.Logf("Generated script:\n%s", script)
}

func TestSSHConnectInvalidHost(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	exec := &SSHExecutor{}
	err := exec.Execute(ctx, "192.0.2.1", 22, "root", "invalid",
		[]string{"echo hello"}, 30*time.Second)

	if err == nil {
		t.Error("expected error for invalid host, got nil")
	}
	t.Logf("expected error: %v", err)
}

func TestSSHConnectTimeout(t *testing.T) {
	// Use a non-routable IP to trigger connection timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	exec := &SSHExecutor{}
	err := exec.Execute(ctx, "10.255.255.1", 22, "root", "password",
		[]string{"echo hello"}, 30*time.Second)

	if err == nil {
		t.Error("expected timeout error, got nil")
	}
	t.Logf("expected timeout/error: %v", err)
}
