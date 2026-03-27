package logger

import (
	"fmt"
	"log"
	"os"
)

var l *log.Logger

// Init opens (or creates) the log file and sets up the package-level logger.
func Init(path string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("open log file %s: %w", path, err)
	}
	l = log.New(f, "", log.LstdFlags)
	return nil
}

// Errorf logs a formatted error message.
func Errorf(format string, args ...any) {
	if l != nil {
		l.Printf("ERROR "+format, args...)
	}
}

// Infof logs a formatted info message.
func Infof(format string, args ...any) {
	if l != nil {
		l.Printf("INFO  "+format, args...)
	}
}
