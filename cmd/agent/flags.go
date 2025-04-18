package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

var (
	serverHost     string
	reportInterval int
	pollInterval   int

	flagHashKey string

	flagRateLimit   int
	flagReportBatch bool

	// Флаги линковщика
	buildVersion string
	buildDate    string
	buildCommit  string

	cryptoKeyPath string

	configPath string
)

func parseFlags() {
	flag.StringVar(&configPath, "c", "", "json config path")
	flag.StringVar(&serverHost, "a", "localhost:8080", "endpoint address")
	flag.IntVar(&reportInterval, "r", 10, "report interval")
	flag.IntVar(&pollInterval, "p", 2, "poll interval")
	flag.StringVar(&flagHashKey, "k", "", "hash key")
	flag.IntVar(&flagRateLimit, "l", 0, "http requests rate limit")
	flag.BoolVar(&flagReportBatch, "b", true, "determinate batch reporting")
	flag.StringVar(&cryptoKeyPath, "crypto-key", "", "path to public rsa crypto key")
	flag.Parse()
}

type Config struct {
	Address        string `json:"address"`
	ReportInterval string `json:"report_interval"`
	PollInterval   string `json:"poll_interval"`
	CryptoKey      string `json:"crypto_key"`
}

func parseJSON() {
	if configPath == "" {
		return
	}

	bts, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalf("Failed to read json config: %v", err)
	}

	var cfg Config
	if err = json.Unmarshal(bts, &cfg); err != nil {
		log.Fatalf("Failed to unmarshal json config: %v", err)
	}

	if serverHost == "" {
		serverHost = cfg.Address
	}
	if cryptoKeyPath == "" {
		cryptoKeyPath = cfg.CryptoKey
	}

	if reportInterval == 10 && cfg.ReportInterval != "" {
		rp, err := strconv.Atoi(strings.TrimSuffix(cfg.ReportInterval, "s"))
		if err != nil {
			log.Fatalf("Failed to convert store_interval: %v", err)
		}
		reportInterval = rp
	}
	if pollInterval == 2 && cfg.PollInterval != "" {
		st, err := strconv.Atoi(strings.TrimSuffix(cfg.PollInterval, "s"))
		if err != nil {
			log.Fatalf("Failed to convert store_interval: %v", err)
		}
		pollInterval = st
	}
}

func overrideEnv() {
	if envServerHost := os.Getenv("ADDRESS"); envServerHost != "" {
		serverHost = envServerHost
	}

	if envReportInterval := os.Getenv("REPORT_INTERVAL"); envReportInterval != "" {
		val, err := strconv.Atoi(envReportInterval)
		if err == nil {
			reportInterval = val
		}
	}

	if envPollInterval := os.Getenv("POLL_INTERVAL"); envPollInterval != "" {
		val, err := strconv.Atoi(envPollInterval)
		if err == nil {
			pollInterval = val
		}
	}

	if envHashKey := os.Getenv("KEY"); envHashKey != "" {
		flagHashKey = envHashKey
	}

	if envRateLimit := os.Getenv("RATE_LIMIT"); envRateLimit != "" {
		val, err := strconv.Atoi(envRateLimit)
		if err == nil {
			flagRateLimit = val
		}
	}

	if envReportBatch := os.Getenv("REPORT_BATCH"); envReportBatch != "" {
		val, err := strconv.ParseBool(envReportBatch)
		if err == nil {
			flagReportBatch = val
		}
	}

	if envCryptoKeyPath := os.Getenv("CRYPTO_KEY"); envCryptoKeyPath != "" {
		cryptoKeyPath = envCryptoKeyPath
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
