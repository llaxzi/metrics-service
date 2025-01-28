package main

import (
	"flag"
	"os"
	"strconv"
)

var serverHost string
var reportInterval int
var pollInterval int

var flagHashKey string

var flagRateLimit int
var flagReportBatch bool

func parseFlags() {
	flag.StringVar(&serverHost, "a", "localhost:8080", "endpoint address")
	flag.IntVar(&reportInterval, "r", 10, "report interval")
	flag.IntVar(&pollInterval, "p", 2, "poll interval")
	flag.StringVar(&flagHashKey, "k", "", "hash key")
	flag.IntVar(&flagRateLimit, "l", 0, "http requests rate limit")
	flag.BoolVar(&flagReportBatch, "b", true, "determinate batch reporting")
	flag.Parse()

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

}
