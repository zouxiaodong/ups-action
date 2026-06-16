package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLoggerWrite(t *testing.T) {
	tmpDir := t.TempDir()
	logger, err := NewLogger(tmpDir, 30)
	if err != nil {
		t.Fatalf("NewLogger failed: %v", err)
	}
	defer logger.Close()

	logger.Logf("test message %d", 42)
	logger.Logf("another message")

	// Read the log file
	files, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("read log dir failed: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 log file, got %d", len(files))
	}

	data, err := os.ReadFile(filepath.Join(tmpDir, files[0].Name()))
	if err != nil {
		t.Fatalf("read log file failed: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "test message 42") {
		t.Errorf("log missing expected content, got: %s", content)
	}
	if !strings.Contains(content, "another message") {
		t.Errorf("log missing second message, got: %s", content)
	}
}

func TestLoggerFileNameFormat(t *testing.T) {
	tmpDir := t.TempDir()
	logger, err := NewLogger(tmpDir, 30)
	if err != nil {
		t.Fatalf("NewLogger failed: %v", err)
	}
	defer logger.Close()

	files, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("read log dir failed: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 log file, got %d", len(files))
	}

	name := files[0].Name()
	today := time.Now().Format("2006-01-02")
	expected := "ups-action-" + today + ".log"
	if name != expected {
		t.Errorf("expected filename %s, got %s", expected, name)
	}
}

func TestLoggerCleanupOldLogs(t *testing.T) {
	tmpDir := t.TempDir()

	// Create some old log files
	oldDates := []string{"2020-01-01", "2020-06-15", "2020-12-31"}
	for _, d := range oldDates {
		oldFile := filepath.Join(tmpDir, "ups-action-"+d+".log")
		if err := os.WriteFile(oldFile, []byte("old log"), 0644); err != nil {
			t.Fatalf("write old log failed: %v", err)
		}
	}

	// Create a current-ish log file (will be kept)
	recentDate := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	recentFile := filepath.Join(tmpDir, "ups-action-"+recentDate+".log")
	if err := os.WriteFile(recentFile, []byte("recent log"), 0644); err != nil {
		t.Fatalf("write recent log failed: %v", err)
	}

	// NewLogger with maxDays=7 should delete logs older than 7 days
	logger, err := NewLogger(tmpDir, 7)
	if err != nil {
		t.Fatalf("NewLogger failed: %v", err)
	}
	defer logger.Close()

	files, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("read log dir failed: %v", err)
	}

	// Old files (2020) should be deleted; recent file + new file should remain
	if len(files) != 2 {
		t.Errorf("expected 2 log files (recent + today), got %d: %v", len(files), files)
	}
}
