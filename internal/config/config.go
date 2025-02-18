package config

import (
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Logger   LoggerConf
	Storage  StorageConf
	Database DatabaseConf
	HTTP     HTTPConf
}

type LoggerConf struct {
	Level    string // Уровень логирования
	FilePath string // Путь к файлу для записи логов (если пусто, пишет в Stderr)
}

type StorageConf struct {
	Type string // Тип хранилища: "sql" или "in-memory"
}

type DatabaseConf struct {
	Host   string // Хост базы данных
	Port   string // Порт базы данных
	User   string // Пользователь базы данных
	Pwd    string // Пароль базы данных
	DBName string // Имя базы данных
}

type HTTPConf struct {
	Host string // Хост HTTP-сервера
	Port string // Порт HTTP-сервера
}

func NewConfig() *Config {
	return &Config{
		Logger: LoggerConf{
			Level:    "INFO",
			FilePath: "",
		},
		Storage: StorageConf{
			Type: "in-memory",
		},
		HTTP: HTTPConf{
			Host: "0.0.0.0",
			Port: "8080",
		},
	}
}

func LoadConfig(path string) (*Config, error) {
	var cfg Config

	// Чтение файла конфигурации
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Декодирование TOML
	if _, err := toml.Decode(string(data), &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
