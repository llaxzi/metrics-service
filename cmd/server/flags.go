package main

import "flag"

var flagRunAddr string

func parseFlags() {
	flag.StringVar(&flagRunAddr, "a", ":8080", "endpoint address")
	flag.Parse()
}
