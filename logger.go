package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Logger writes to both stdout and a date-based log file.
type Logger struct {
	mu       sync.Mutex
	dir      string
	maxDays  int
	file     *os.File
	currDate string
}

// NewLogger creates a Logger that writes to dir/ups-action-YYYY-MM-DD.log.
// Old logs beyond maxDays are deleted on creation.
func NewLogger(dir string, maxDays int) (*Logger, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory %s: %w", dir, err)
	}

	// Cleanup old logs
	cleanupOldLogs(dir, maxDays)

	l := &Logger{
		dir:     dir,
		maxDays: maxDays,
	}

	if err := l.rotate(); err != nil {
		return nil, err
	}

	return l, nil
}

// Logf writes a line to the log file and stdout, with a timestamp prefix.
func (l *Logger) Logf(format string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()

	// Rotate if date changed
	today := now.Format("2006-01-02")
	if l.currDate != today {
		l.rotate()
		// Also cleanup old logs on date change
		cleanupOldLogs(l.dir, l.maxDays)
	}

	line := now.Format("2006-01-02 15:04:05") + " " + fmt.Sprintf(format, args...)
	if !strings.HasSuffix(line, "\n") {
		line += "\n"
	}

	// Write to stdout
	os.Stdout.WriteString(line)

	// Write to file
	if l.file != nil {
		l.file.WriteString(line)
		l.file.Sync()
	}
}

// Close closes the log file.
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file != nil {
		err := l.file.Close()
		l.file = nil
		return err
	}
	return nil
}

func (l *Logger) rotate() error {
	if l.file != nil {
		l.file.Close()
	}

	today := time.Now().Format("2006-01-02")
	filename := filepath.Join(l.dir, "ups-action-"+today+".log")

	f, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file %s: %w", filename, err)
	}

	l.file = f
	l.currDate = today
	return nil
}

func cleanupOldLogs(dir string, maxDays int) {
	cutoff := time.Now().AddDate(0, 0, -maxDays)

	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasPrefix(name, "ups-action-") || !strings.HasSuffix(name, ".log") {
			continue
		}

		// Parse date from filename: ups-action-2006-01-02.log
		dateStr := strings.TrimPrefix(name, "ups-action-")
		dateStr = strings.TrimSuffix(dateStr, ".log")

		t, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			continue
		}

		if t.Before(cutoff) {
			os.Remove(filepath.Join(dir, name))
		}
	}
}
