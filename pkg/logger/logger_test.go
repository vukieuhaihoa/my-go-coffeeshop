package logger

import (
	"bytes"
	"io"
	"log/slog"
	"strings"
	"testing"
)

func TestParseLevel(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    slog.Level
		wantErr bool
	}{
		{"debug level", "debug", slog.LevelDebug, false},
		{"info level", "info", slog.LevelInfo, false},
		{"warn level", "warn", slog.LevelWarn, false},
		{"warning level", "warning", slog.LevelWarn, false},
		{"error level", "error", slog.LevelError, false},
		{"upper case", "WARN", slog.LevelWarn, false},
		{"typo from reference", "warm", 0, true},
		{"unknown level", "unknown", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseLevel(tt.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseLevel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseLevel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewSlogLogger(t *testing.T) {
	tests := []struct {
		name    string
		level   string
		wantErr bool
	}{
		{"valid level debug", "debug", false},
		{"valid level info", "info", false},
		{"valid level warn", "warn", false},
		{"valid level error", "error", false},
		{"invalid level", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := NewSlogLogger(io.Discard, tt.level)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewSlogLogger() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if logger == nil && !tt.wantErr {
				t.Errorf("NewSlogLogger() returned nil logger for valid level %s", tt.level)
			}
		})
	}
}

func TestSlogLoggerMethods(t *testing.T) {
	var buf bytes.Buffer
	logger, err := NewSlogLogger(&buf, "debug")
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	tests := []struct {
		name string
		fn   func(string, ...any)
	}{
		{"Debug", logger.Debug},
		{"Info", logger.Info},
		{"Warn", logger.Warn},
		{"Error", logger.Error},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("%s() panicked: %v", tt.name, r)
				}
			}()
			tt.fn("test message")
		})
	}
}

func TestSlogLogger_FilterByLevel(t *testing.T) {
	tests := []struct {
		name      string
		logLevel  string
		wantDebug bool
		wantInfo  bool
		wantWarn  bool
		wantError bool
	}{
		{"debug level", "debug", true, true, true, true},
		{"info level", "info", false, true, true, true},
		{"warn level", "warn", false, false, true, true},
		{"error level", "error", false, false, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger, err := NewSlogLogger(&buf, tt.logLevel)
			if err != nil {
				t.Fatalf("Failed to create logger: %v", err)
			}

			logger.Debug("msg-debug")
			logger.Info("msg-info")
			logger.Warn("msg-warn")
			logger.Error("msg-error")

			logOutput := buf.String()

			want := map[string]bool{
				"msg-debug": tt.wantDebug,
				"msg-info":  tt.wantInfo,
				"msg-warn":  tt.wantWarn,
				"msg-error": tt.wantError,
			}

			for msg, shouldExist := range want {
				if shouldExist && !strings.Contains(logOutput, msg) {
					t.Errorf("Expected message %q to be logged at level %s, but not found in output: %s", msg, tt.logLevel, logOutput)
				}
				if !shouldExist && strings.Contains(logOutput, msg) {
					t.Errorf("Message %q should not be logged at level %s, but found in output: %s", msg, tt.logLevel, logOutput)
				}
			}
		})
	}
}
