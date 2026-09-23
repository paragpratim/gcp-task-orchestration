package logger

import (
	"context"
	"log/slog"
	"os"
	"runtime"
	"strings"
	"time"
)

// Init configures the global structured logger.
func Init() {
	level := getLogLevel()
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
	})
	slog.SetDefault(slog.New(handler))
}

func getLogLevel() slog.Level {
	switch strings.ToLower(os.Getenv("LOG_LEVEL")) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func logWithCaller(level slog.Level, msg string, args ...any) {
	logger := slog.Default()
	if !logger.Enabled(context.Background(), level) {
		return
	}

	var pcs [1]uintptr
	// Skip 3 frames: runtime.Callers -> logWithCaller -> logger.Info -> Real caller
	runtime.Callers(3, pcs[:])

	record := slog.NewRecord(time.Now(), level, msg, pcs[0])
	record.Add(args...)
	_ = logger.Handler().Handle(context.Background(), record)
}

// Info logs a message at INFO level.
func Info(msg string, args ...any) {
	logWithCaller(slog.LevelInfo, msg, args...)
}

// Warn logs a message at WARN level.
func Warn(msg string, args ...any) {
	logWithCaller(slog.LevelWarn, msg, args...)
}

// Error logs a message at ERROR level.
func Error(msg string, args ...any) {
	logWithCaller(slog.LevelError, msg, args...)
}

// Debug logs a message at DEBUG level.
func Debug(msg string, args ...any) {
	logWithCaller(slog.LevelDebug, msg, args...)
}

func Fatal(msg string, args ...any) {
	logWithCaller(slog.LevelError, msg, args...)
	os.Exit(1)
}
