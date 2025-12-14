package main

import (
	"central-logs/internal/config"
)

// envhelp is a utility to show all supported environment variables
func main() {
	config.PrintEnvHelp()
}
