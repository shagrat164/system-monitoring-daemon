package logger

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/shagrat164/system-monitoring-daemon/internal/config"
	"github.com/stretchr/testify/require"
)

// TestNewLogger проверяет создание логгера с разными уровнями логирования.
func TestNewLogger(t *testing.T) {
	tests := []struct {
		name          string
		level         string
		expectedLevel string
	}{
		{"Debug Level", "debug", constLevelDebug},
		{"Info Level", "info", constLevelInfo},
		{"Warn Level", "warn", constLevelWarn},
		{"Error Level", "error", constLevelError},
		{"Invalid Level", "unknown", constLevelInfo}, // По умолчанию INFO
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.LoggerConfig{Level: tt.level}
			logger, err := New(cfg)
			require.NoError(t, err)
			require.Equal(t, tt.expectedLevel, logger.level)
		})
	}
}

// TestLogLevels проверяет, что сообщения логируются только на допустимых уровнях.
func TestLogLevels(t *testing.T) {
	tests := []struct {
		name        string
		level       string
		debugLogged bool
		infoLogged  bool
		warnLogged  bool
		errorLogged bool
	}{
		{"Debug Level", constLevelDebug, true, true, true, true},
		{"Info Level", constLevelInfo, false, true, true, true},
		{"Warn Level", constLevelWarn, false, false, true, true},
		{"Error Level", constLevelError, false, false, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			cfg := config.LoggerConfig{Level: tt.level}
			logger, err := New(cfg)
			require.NoError(t, err)
			logger.logWriter.SetOutput(&buf)

			logger.Debug("debug message")
			logger.Info("info message")
			logger.Warn("warn message")
			logger.Error("error message")

			output := buf.String()
			if tt.debugLogged {
				require.Contains(t, output, `"level":"DEBUG"`)
			} else {
				require.NotContains(t, output, `"level":"DEBUG"`)
			}
			if tt.infoLogged {
				require.Contains(t, output, `"level":"INFO"`)
			} else {
				require.NotContains(t, output, `"level":"INFO"`)
			}
			if tt.warnLogged {
				require.Contains(t, output, `"level":"WARN"`)
			} else {
				require.NotContains(t, output, `"level":"WARN"`)
			}
			if tt.errorLogged {
				require.Contains(t, output, `"level":"ERROR"`)
			} else {
				require.NotContains(t, output, `"level":"ERROR"`)
			}
		})
	}
}

// TestLogFormat проверяет формат вывода логов (JSON).
func TestLogFormat(t *testing.T) {
	var buf bytes.Buffer
	cfg := config.LoggerConfig{Level: constLevelInfo}
	logger, err := New(cfg)
	require.NoError(t, err)
	logger.logWriter.SetOutput(&buf)

	logger.Info("test message")

	var entry logEntry
	err = json.Unmarshal(buf.Bytes(), &entry)
	require.NoError(t, err)
	require.Equal(t, constLevelInfo, entry.Level)
	require.Equal(t, "test message", entry.Message)
	require.NotEmpty(t, entry.Timestamp)

	// Проверяем, что timestamp в формате RFC3339
	_, err = time.Parse(time.RFC3339, entry.Timestamp)
	require.NoError(t, err)
}

// TestLogToFile проверяет запись логов в файл.
func TestLogToFile(t *testing.T) {
	// Создаем временный файл для логов
	tmpFile, err := os.CreateTemp("", "test-log-*.log")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	cfg := config.LoggerConfig{
		Level: constLevelInfo,
		Path:  tmpFile.Name(),
	}
	logger, err := New(cfg)
	require.NoError(t, err)

	logger.Info("test message")

	// Читаем содержимое файла
	content, err := os.ReadFile(tmpFile.Name())
	require.NoError(t, err)

	var entry logEntry
	err = json.Unmarshal(content, &entry)
	require.NoError(t, err)
	require.Equal(t, constLevelInfo, entry.Level)
	require.Equal(t, "test message", entry.Message)
	require.NotEmpty(t, entry.Timestamp)
}
