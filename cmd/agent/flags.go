package main

import "flag"

var serverHost string
var reportInterval int
var pollInterval int

func parseFlags() {
	flag.StringVar(&serverHost, "a", "localhost:8080", "endpoint address")
	flag.IntVar(&reportInterval, "r", 10, "report interval")
	flag.IntVar(&pollInterval, "p", 2, "poll interval")
	flag.Parse()
}
