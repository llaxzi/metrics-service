package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

var (
	flagRunAddr  string
	flagLogLevel string

	flagStoreInterval   int
	flagFileStoragePath string
	flagRestore         bool

	flagDatabaseDSN string

	flagHashKey string

	// Флаги линковщика
	buildVersion string
	buildDate    string
	buildCommit  string
)

func parseFlags() {
	flag.StringVar(&flagRunAddr, "a", "localhost:8080", "endpoint address")
	flag.StringVar(&flagLogLevel, "l", "info", "log level")

	flag.IntVar(&flagStoreInterval, "i", 300, "metrics store interval")
	flag.StringVar(&flagFileStoragePath, "f", "metrics.json", "metrics store path")
	flag.BoolVar(&flagRestore, "r", true, "load metrics bool")

	flag.StringVar(&flagDatabaseDSN, "d", "", "database DSN")

	flag.StringVar(&flagHashKey, "k", "", "hash key")

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
	if envHashKey := os.Getenv("KEY"); envHashKey != "" {
		flagHashKey = envHashKey
	}
}

func printBuildInfo() {

	buildVersion = filterFlag(buildVersion)
	buildDate = filterFlag(buildDate)
	buildCommit = filterFlag(buildCommit)

	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %v\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
}

func filterFlag(flag string) string {
	if flag == "" {
		return "N/A"
	}
	return flag
}
