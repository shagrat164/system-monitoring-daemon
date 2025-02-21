package main

import (
	"flag"
	"log"

	"github.com/shagrat164/system-monitoring-daemon/internal/server"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "/etc/calendar/config.yaml", "Path to configuration file")
}

func main() {
	flag.Parse()

	if flag.Arg(0) == "version" {
		printVersion()
		return
	}

	const port = 50051

	log.Println("Starting system monitoring daemon...")
	if err := server.Run(port); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
