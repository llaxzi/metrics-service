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

	cryptoKeyPath string

	configPath string
)

func parseFlags() {
	flag.StringVar(&configPath, "c", "", "json config path")

	flag.StringVar(&flagRunAddr, "a", ":8080", "endpoint address")
	flag.StringVar(&flagLogLevel, "l", "", "log level")

	flag.IntVar(&flagStoreInterval, "i", 300, "metrics store interval")
	flag.StringVar(&flagFileStoragePath, "f", "", "metrics store path")
	flag.BoolVar(&flagRestore, "r", false, "load metrics bool")

	flag.StringVar(&flagDatabaseDSN, "d", "", "database DSN")
	flag.StringVar(&flagHashKey, "k", "", "hash key")
	flag.StringVar(&cryptoKeyPath, "crypto-key", "", "path to private rsa crypto key")

	flag.Parse()
}

type Config struct {
	Address       string `json:"address"`
	Restore       bool   `json:"restore"`
	StoreInterval string `json:"store_interval"`
	DatabaseDSN   string `json:"database_dsn"`
	CryptoKey     string `json:"crypto_key"`
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

	if flagRunAddr == "" {
		flagRunAddr = cfg.Address
	}
	if flagDatabaseDSN == "" {
		flagDatabaseDSN = cfg.DatabaseDSN
	}
	if cryptoKeyPath == "" {
		cryptoKeyPath = cfg.CryptoKey
	}
	if flagStoreInterval == 300 && cfg.StoreInterval != "" {
		storeInterval, err := strconv.Atoi(strings.TrimSuffix(cfg.StoreInterval, "s"))
		if err != nil {
			log.Fatalf("Failed to convert store_interval: %v", err)
		}
		flagStoreInterval = storeInterval
	}
	// флаг -r по умолчанию false, не трогаем если он true
	if !flagRestore {
		flagRestore = cfg.Restore
	}
}

func overrideEnv() {
	if env := os.Getenv("ADDRESS"); env != "" {
		flagRunAddr = env
	}
	if env := os.Getenv("LOG_LEVEL"); env != "" {
		flagLogLevel = env
	}
	if env := os.Getenv("STORE_INTERVAL"); env != "" {
		if i, err := strconv.Atoi(env); err == nil {
			flagStoreInterval = i
		}
	}
	if env := os.Getenv("FILE_STORAGE_PATH"); env != "" {
		flagFileStoragePath = env
	}
	if env := os.Getenv("RESTORE"); env != "" {
		if b, err := strconv.ParseBool(env); err == nil {
			flagRestore = b
		}
	}
	if env := os.Getenv("DATABASE_DSN"); env != "" {
		flagDatabaseDSN = env
	}
	if env := os.Getenv("KEY"); env != "" {
		flagHashKey = env
	}
	if env := os.Getenv("CRYPTO_KEY"); env != "" {
		cryptoKeyPath = env
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
