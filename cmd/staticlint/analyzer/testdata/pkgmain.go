package main

import "os"

func main() {
	os.Exit(1) // want "call to os.Exit in main function"
}
