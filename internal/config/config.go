package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/BurntSushi/toml"
)

// Config структура конфигурации демона.
type Config struct {
	GRPCPort string        `toml:"grpc_port"` // Порт gRPC-сервера
	Logger   LoggerConfig  `toml:"logger"`    // Конфигурация логгера
	Enabled  MetricsConfig `toml:"metrics"`   // Включенные подсистемы
}

// LoggerConfig структура конфигурации логгера.
type LoggerConfig struct {
	Level string `toml:"level"` // Уровень логирования (DEBUG, INFO, WARN, ERROR)
	Path  string `toml:"path"`  // Путь к файлу логов (пусто = stderr)
}

// MetricsConfig флаги включения/выключения подсистем.
type MetricsConfig struct {
	LoadAvg    bool `toml:"load_avg"`   // Сбор load average
	CPU        bool `toml:"cpu"`        // Сбор информации о ЦПУ
	Disk       bool `toml:"disk"`       // Сбор информации о дисках
	Filesystem bool `toml:"filesystem"` // Сбор информации о файловых системах
}

// NewConfig создает конфигурацию по умолчанию.
func NewConfig() *Config {
	return &Config{
		Logger: LoggerConfig{
			Level: "info",
			Path:  "",
		},
		GRPCPort: "50051", // Порт по умолчанию
		Enabled: MetricsConfig{
			LoadAvg: true, // По умолчанию включен только load average
		},
	}
}

// LoadConfig загружает конфигурацию из TOML-файла и аргументов командной строки.
func LoadConfig(path, port string) (*Config, error) {
	cfg := NewConfig()

	// Если конфиг-файл указан, читаем его
	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		if _, err := toml.Decode(string(data), cfg); err != nil {
			return nil, err
		}
	}

	// Если порт указан, минимальная прооверка на корректность и запись в конфиг
	if port != "" {
		p, err := strconv.Atoi(port)
		if err != nil {
			fmt.Fprintf(os.Stderr, "port is incorrect: %v\n", err)
			port = cfg.GRPCPort
		}
		if p < 1000 || p > 65535 {
			port = cfg.GRPCPort
		}
		cfg.GRPCPort = port
	}

	return cfg, nil
}
