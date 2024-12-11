package main

import (
	"flag"
	"os"
)

var flagRunAddr string
var flagLogLevel string

func parseFlags() {
	flag.StringVar(&flagRunAddr, "a", "localhost:8080", "endpoint address")
	flag.StringVar(&flagLogLevel, "l", "info", "log level")
	flag.Parse()

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		flagRunAddr = envRunAddr
	}
	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		flagLogLevel = envLogLevel
	}
}
