package logger

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/shagrat164/system-monitoring-daemon/internal/config"
)

// Уровни логирования.
const (
	constLevelDebug = "DEBUG" // 10
	constLevelInfo  = "INFO"  // 20
	constLevelWarn  = "WARN"  // 30
	constLevelError = "ERROR" // 40
)

type Logger struct {
	level     string
	logWriter *log.Logger
}

func New(cfg config.LoggerConfig) (*Logger, error) {
	// Настройка выходного потока (Stderr или файл)
	var output *os.File
	if cfg.Path != "" {
		var err error
		output, err = os.OpenFile(cfg.Path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
	} else {
		output = os.Stderr
	}

	// Проверка корректного уровня логирования
	level := strings.ToUpper(cfg.Level)
	if level != constLevelInfo && level != constLevelDebug && level != constLevelWarn && level != constLevelError {
		level = constLevelInfo // Info default
	}

	return &Logger{
		level:     level,
		logWriter: log.New(output, "", 0),
	}, nil
}

// logEntry структура для JSON-логов.
type logEntry struct {
	Level     string `json:"level"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

// log выводит сообщение msg на заданном уровне level.
func (l *Logger) log(level, msg string) {
	if l.shouldLog(level) {
		entry := logEntry{
			Level:     level,
			Message:   msg,
			Timestamp: time.Now().Format(time.RFC3339),
		}
		data, _ := json.Marshal(entry)
		l.logWriter.Println(string(data))
	}
}

// shouldLog проверяет, требуется ли логировать сообщение на заданном уровне level.
func (l *Logger) shouldLog(level string) bool {
	levels := map[string]int{
		constLevelDebug: 10,
		constLevelInfo:  20,
		constLevelWarn:  30,
		constLevelError: 40,
	}
	return levels[level] >= levels[l.level]
}

// Debug выводит сообщение уровня DEBUG.
func (l *Logger) Debug(msg string) {
	l.log(constLevelDebug, msg)
}

// Info выводит сообщение уровня INFO.
func (l *Logger) Info(msg string) {
	l.log(constLevelInfo, msg)
}

// Warn выводит сообщение уровня WARN.
func (l *Logger) Warn(msg string) {
	l.log(constLevelWarn, msg)
}

// Error выводит сообщение уровня ERROR.
func (l *Logger) Error(msg string) {
	l.log(constLevelError, msg)
}
