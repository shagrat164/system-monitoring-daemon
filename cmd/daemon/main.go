package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/shagrat164/system-monitoring-daemon/internal/config"
	"github.com/shagrat164/system-monitoring-daemon/internal/logger"
	"github.com/shagrat164/system-monitoring-daemon/internal/server"
)

var (
	configFile string
	port       string
)

func init() {
	flag.StringVar(&configFile, "config", "", "Path to configuration file")
	flag.StringVar(&port, "port", "", "gRPC server port")
}

func main() {
	flag.Parse()

	if flag.Arg(0) == "version" {
		printVersion()
		return
	}

	// Загрузка конфигурации
	cfg, err := config.LoadConfig(configFile, port)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		fmt.Fprintf(os.Stderr, "Load default config.")
		cfg = config.NewConfig()
	}

	// Инициализация логгера
	log, err := logger.New(cfg.Logger)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	log.Info("Starting system monitoring daemon")

	// Запуск периодического обновления конфигурации
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			updateConfig, err := config.LoadConfig(configFile, port)
			if err != nil {
				log.Error(fmt.Sprintf("Configuration update error: %v\n", err))
				continue
			}
			cfg = updateConfig
			log.Info("Configuration has been updated")
		}
	}()

	// Запуск gRPC-сервера
	if err := server.Run(cfg, log); err != nil {
		log.Error(fmt.Sprintf("Server failed: %v", err))
		os.Exit(1)
	}
}
