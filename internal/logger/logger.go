package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
	"time"
)

var (
	debugLog *os.File
	logPath  string
)

// Init initializes the debug logger
func Init() error {
	// Create log directory in user's home
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	logDir := filepath.Join(homeDir, ".fight-the-landlord")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Create or append to debug.log
	logPath = filepath.Join(logDir, "debug.log")
	debugLog, err = os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	// Rotate if file is too large (> 10MB)
	if info, err := debugLog.Stat(); err == nil && info.Size() > 10*1024*1024 {
		_ = debugLog.Close()
		backupPath := filepath.Join(logDir, fmt.Sprintf("debug.log.%d", time.Now().Unix()))
		_ = os.Rename(logPath, backupPath)
		debugLog, err = os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			return fmt.Errorf("failed to create new log file: %w", err)
		}
	}

	// Set log output to file
	log.SetOutput(debugLog)
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	LogInfo("Logger initialized, log file: %s", logPath)
	return nil
}

// Close closes the debug log file
func Close() {
	if debugLog != nil {
		_ = debugLog.Close()
	}
}

// LogInfo logs an info message
func LogInfo(format string, args ...interface{}) {
	log.Printf("[INFO] "+format, args...)
}

// LogError logs an error message
func LogError(format string, args ...interface{}) {
	log.Printf("[ERROR] "+format, args...)
}

// LogPanic logs a panic with stack trace
func LogPanic(r interface{}) {
	log.Printf("[PANIC] %v\n%s", r, debug.Stack())
}

// GetLogPath returns the current log file path
func GetLogPath() string {
	return logPath
}
