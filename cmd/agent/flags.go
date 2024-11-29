package main

import "flag"

var reportInterval int
var pollInterval int

func parseFlags() {
	flag.IntVar(&reportInterval, "r", 10, "report interval")
	flag.IntVar(&pollInterval, "p", 2, "poll interval")

}
