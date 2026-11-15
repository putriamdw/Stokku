package config

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/natefinch/lumberjack.v2"
)

// NewLogger menulis log JSON ke layar dan logs/app.log
func NewLogger() *slog.Logger {
	if err := os.MkdirAll("logs", 0o755); err != nil {
		panic("gagal membuat folder logs: " + err.Error())
	}
	rotator := &lumberjack.Logger{
		Filename:   filepath.Join("logs", "app.log"),
		MaxSize:    10,
		MaxBackups: 5,
		MaxAge:     14,
		Compress:   true,
	}
	out := io.MultiWriter(os.Stdout, rotator)
	jsonLog := slog.NewJSONHandler(out, &slog.HandlerOptions{
		Level: parseLevel(GetEnv("LOG_LEVEL", "info")),
	})
	logger := slog.New(jsonLog)
	slog.SetDefault(logger)
	return logger
}

func parseLevel(value string) slog.Level {
	switch strings.ToLower(value) {
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
