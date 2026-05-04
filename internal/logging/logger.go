package logging

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	logDirName     = "logs"
	activeLogName  = "amr-disp.log"
	archivePattern = "amr-disp-%s.log"
	timeFormat     = "2006-01-02-150405"
)

func New(workspace, level string) (*slog.Logger, io.Closer, string, error) {
	logsDir := filepath.Join(workspace, logDirName)
	if err := os.MkdirAll(logsDir, 0o755); err != nil {
		return nil, nil, "", fmt.Errorf("create logs dir: %w", err)
	}

	activePath := filepath.Join(logsDir, activeLogName)
	if err := rotate(activePath, time.Now()); err != nil {
		return nil, nil, "", fmt.Errorf("rotate active log: %w", err)
	}

	file, err := os.OpenFile(activePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, nil, "", fmt.Errorf("open active log: %w", err)
	}

	logLevel, err := ParseLevel(level)
	if err != nil {
		_ = file.Close()
		return nil, nil, "", err
	}

	handler := slog.NewTextHandler(io.MultiWriter(os.Stderr, file), &slog.HandlerOptions{
		Level: logLevel,
	})
	return slog.New(handler).With("app", "amr-disp"), file, activePath, nil
}

func ParseLevel(value string) (slog.Level, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "info":
		return slog.LevelInfo, nil
	case "debug":
		return slog.LevelDebug, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("unsupported log level %q", value)
	}
}

func rotate(activePath string, now time.Time) error {
	info, err := os.Stat(activePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("active log path %s is a directory", activePath)
	}

	dir := filepath.Dir(activePath)
	base := fmt.Sprintf(archivePattern, now.Local().Format(timeFormat))
	archivedPath := filepath.Join(dir, base)
	if _, err := os.Stat(archivedPath); err == nil {
		archivedPath = filepath.Join(dir, fmt.Sprintf("amr-disp-%s-%d.log", now.Local().Format(timeFormat), now.UnixNano()))
	} else if !os.IsNotExist(err) {
		return err
	}
	return os.Rename(activePath, archivedPath)
}

func DiscardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
