package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// SSHExecutor manages SSH connections and command execution.
type SSHExecutor struct{}

// Execute connects to the remote host via SSH with password auth,
// executes the given script via bash -s, and returns the combined output.
func (e *SSHExecutor) Execute(ctx context.Context, host string, port int,
	user, password string, script []string, execTimeout time.Duration) error {

	scriptContent := buildScript(script)

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", host, port)

	// Dial with context timeout
	var client *ssh.Client
	var err error
	done := make(chan struct{})

	go func() {
		defer close(done)
		client, err = ssh.Dial("tcp", addr, config)
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("SSH connect to %s timeout: %w", addr, ctx.Err())
	case <-done:
		if err != nil {
			return fmt.Errorf("SSH dial %s failed: %w", addr, err)
		}
	}

	defer client.Close()

	return e.execWithTimeout(ctx, client, scriptContent, execTimeout)
}

func (e *SSHExecutor) execWithTimeout(ctx context.Context, client *ssh.Client,
	scriptContent string, timeout time.Duration) error {

	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

	// Get stdout and stderr for logging
	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	// Pipe script to bash via stdin
	stdin, err := session.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	// Start bash -s (reads script from stdin)
	if err := session.Start("bash -s"); err != nil {
		return fmt.Errorf("failed to start bash -s: %w", err)
	}

	// Write script to stdin
	io.WriteString(stdin, scriptContent)
	stdin.Close()

	// Wait with timeout
	errCh := make(chan error, 1)
	go func() {
		errCh <- session.Wait()
	}()

	select {
	case <-ctx.Done():
		session.Signal(ssh.SIGKILL)
		return fmt.Errorf("execution timeout: %w (output: %s%s)",
			ctx.Err(), stdout.String(), stderr.String())
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("script execution failed: %w\nstdout: %s\nstderr: %s",
				err, stdout.String(), stderr.String())
		}
		// Log successful output
		fmt.Printf("[SSH output]\n%s", stdout.String())
		if stderr.Len() > 0 {
			fmt.Printf("[SSH stderr]\n%s", stderr.String())
		}
	}
	return nil
}

func buildScript(commands []string) string {
	var sb strings.Builder
	sb.WriteString("#!/bin/bash\n")
	sb.WriteString("set -e\n")
	sb.WriteString("set -o pipefail\n")
	sb.WriteString("\n")
	sb.WriteString("# UPS Crisis Shutdown Script\n")
	sb.WriteString("# Executed at: $(date)\n")
	sb.WriteString("\n")
	for _, cmd := range commands {
		sb.WriteString("echo '>>> Executing: ")
		sb.WriteString(cmd)
		sb.WriteString("'\n")
		sb.WriteString(cmd)
		sb.WriteString("\n")
		sb.WriteString("echo '>>> Done: ")
		sb.WriteString(cmd)
		sb.WriteString("'\n")
		sb.WriteString("\n")
	}
	sb.WriteString("echo '>>> All commands completed'\n")
	return sb.String()
}
