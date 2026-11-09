// Package logger provides a simple logging interface and an implementation using slog.
package logger

import (
	"fmt"
	"io"
	"log/slog"
	"strings"
)

// Logger is an interface that defines the logging methods.
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

// SlogLogger is an implementation of the Logger interface using slog.
type SlogLogger struct {
	l *slog.Logger
}

var _ Logger = (*SlogLogger)(nil)

// NewSlogLogger creates a new SlogLogger instance with the specified writer and log level.
func NewSlogLogger(w io.Writer, level string) (*SlogLogger, error) {
	lvl, err := parseLevel(level)
	if err != nil {
		return nil, err
	}
	slogger := slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{Level: lvl}))
	return &SlogLogger{
		l: slogger,
	}, nil
}

func parseLevel(s string) (slog.Level, error) {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("unknown log level %q, valid levels are: debug, info, warn, error", s)
	}
}

// Debug logs a message at the debug level.
func (s *SlogLogger) Debug(msg string, args ...any) {
	s.l.Debug(msg, args...)
}

// Info logs a message at the info level.
func (s *SlogLogger) Info(msg string, args ...any) {
	s.l.Info(msg, args...)
}

// Warn logs a message at the warn level.
func (s *SlogLogger) Warn(msg string, args ...any) {
	s.l.Warn(msg, args...)
}

// Error logs a message at the error level.
func (s *SlogLogger) Error(msg string, args ...any) {
	s.l.Error(msg, args...)
}
