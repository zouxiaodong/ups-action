package main

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"
)

// ServerResult holds the outcome for a single server.
type ServerResult struct {
	Host    string `json:"host"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// IsSuccess returns true if the server execution was successful.
func (r *ServerResult) IsSuccess() bool {
	return r.Success
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <config.json>\n", os.Args[0])
		os.Exit(3)
	}

	configPath := os.Args[1]

	cfg, err := ParseConfigFile(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(3)
	}
	cfg.FillDefaults()

	// Initialize logger
	logger, err := NewLogger(cfg.Global.LogDir, cfg.Global.LogMaxDays)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(3)
	}
	defer logger.Close()

	logger.Logf("UPS Crisis Action Tool started")
	logger.Logf("Config loaded: %d server(s) configured", len(cfg.Servers))

	if len(cfg.Servers) == 0 {
		logger.Logf("No servers configured, nothing to do. Exiting.")
		os.Exit(0)
	}

	// Run all servers in parallel
	results := runAllServers(cfg)

	// Log summary
	logger.Logf("=== Summary ===")
	successCount := 0
	for _, r := range results {
		if r.Success {
			successCount++
			logger.Logf("  [OK]    %s", r.Host)
		} else {
			logger.Logf("  [FAIL]  %s: %s", r.Host, r.Error)
		}
	}
	logger.Logf("%d/%d servers succeeded", successCount, len(results))

	if allSucceeded(results) {
		logger.Logf("All servers succeeded, exiting with code 0")
		os.Exit(0)
	} else {
		logger.Logf("Some servers failed, exiting with code 1")
		os.Exit(1)
	}
}

func runAllServers(cfg *Config) []ServerResult {
	var wg sync.WaitGroup
	resultsCh := make(chan ServerResult, len(cfg.Servers))

	for _, srv := range cfg.Servers {
		wg.Add(1)
		go func(s ServerConfig) {
			defer wg.Done()
			result := executeServer(cfg, s)
			resultsCh <- result
		}(srv)
	}

	wg.Wait()
	close(resultsCh)

	results := make([]ServerResult, 0, len(cfg.Servers))
	for r := range resultsCh {
		results = append(results, r)
	}
	return results
}

func executeServer(cfg *Config, srv ServerConfig) ServerResult {
	result := ServerResult{Host: srv.Host}

	log := os.Stdout // Use stdout since we don't have logger access in this goroutine
	fmt.Fprintf(log, "[%s] Starting shutdown sequence...\n", srv.Host)

	executor := &SSHExecutor{}

	// Create a context that combines connect + execute timeouts
	totalTimeout := time.Duration(srv.ConnectTimeoutSeconds+srv.ExecuteTimeoutSeconds+10) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), totalTimeout)
	defer cancel()

	// Retry logic: retry connecting and executing the script
	execTimeout := time.Duration(srv.ExecuteTimeoutSeconds) * time.Second
	retryDelay := time.Duration(cfg.Global.RetryDelaySeconds) * time.Second

	err := Retry(ctx, cfg.Global.RetryCount+1, retryDelay, func() error {
		// Each attempt gets its own context with the execute timeout
		attemptCtx, attemptCancel := context.WithTimeout(ctx, execTimeout+time.Duration(srv.ConnectTimeoutSeconds)*time.Second)
		defer attemptCancel()

		return executor.Execute(attemptCtx, srv.Host, srv.Port, srv.User, srv.Password,
			srv.Script, execTimeout)
	})

	if err != nil {
		result.Error = err.Error()
		fmt.Fprintf(log, "[%s] FAILED: %v\n", srv.Host, err)
	} else {
		result.Success = true
		fmt.Fprintf(log, "[%s] SUCCESS\n", srv.Host)
	}

	return result
}

func allSucceeded(results []ServerResult) bool {
	for _, r := range results {
		if !r.Success {
			return false
		}
	}
	return true
}
