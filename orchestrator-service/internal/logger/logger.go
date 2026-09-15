package logger

import (
	"log/slog"
	"os"
)

// Init configures the global structured logger.
// Call this once from main() before any other logging calls.
func Init() {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	slog.SetDefault(slog.New(handler))
}

// Info logs a message at INFO level.
func Info(msg string, args ...any) {
	slog.Info(msg, args...)
}

// Warn logs a message at WARN level.
func Warn(msg string, args ...any) {
	slog.Warn(msg, args...)
}

// Error logs a message at ERROR level.
func Error(msg string, args ...any) {
	slog.Error(msg, args...)
}

// Debug logs a message at DEBUG level.
func Debug(msg string, args ...any) {
	slog.Debug(msg, args...)
}

func Fatal(msg string, args ...any) {
	slog.Error(msg, args...)
	os.Exit(1)
}
