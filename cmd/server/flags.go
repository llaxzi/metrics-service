package main

import (
	"flag"
	"os"
	"strconv"
)

var flagRunAddr string
var flagLogLevel string

var flagStoreInterval int
var flagFileStoragePath string
var flagRestore bool

var flagDatabaseDSN string

func parseFlags() {
	flag.StringVar(&flagRunAddr, "a", "localhost:8080", "endpoint address")
	flag.StringVar(&flagLogLevel, "l", "info", "log level")

	flag.IntVar(&flagStoreInterval, "i", 300, "metrics store interval")
	flag.StringVar(&flagFileStoragePath, "f", "metrics.json", "metrics store path")
	flag.BoolVar(&flagRestore, "r", true, "load metrics bool")

	flag.StringVar(&flagDatabaseDSN, "d", "", "database DSN")

	flag.Parse()

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		flagRunAddr = envRunAddr
	}
	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		flagLogLevel = envLogLevel
	}
	if envStoreInterval := os.Getenv("STORE_INTERVAL"); envStoreInterval != "" {
		interval, err := strconv.Atoi(envStoreInterval)
		if err == nil {
			flagStoreInterval = interval
		}
	}
	if envFileStoragePath := os.Getenv("LOG_LEVEL"); envFileStoragePath != "" {
		flagLogLevel = envFileStoragePath
	}
	if envRestore := os.Getenv("RESTORE"); envRestore != "" {
		restore, err := strconv.ParseBool(envRestore)
		if err == nil {
			flagRestore = restore
		}
	}
	if envDatabaseDSN := os.Getenv("DATABASE_DSN"); envDatabaseDSN != "" {
		flagDatabaseDSN = envDatabaseDSN
	}

}
