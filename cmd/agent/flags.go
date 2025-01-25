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

func parseFlags() {
	flag.StringVar(&serverHost, "a", "localhost:8080", "endpoint address")
	flag.IntVar(&reportInterval, "r", 10, "report interval")
	flag.IntVar(&pollInterval, "p", 2, "poll interval")
	flag.StringVar(&flagHashKey, "k", "", "hash key")
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

}
