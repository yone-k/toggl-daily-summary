package main

import (
	"os"

	"github.com/yone/toggl-daily-summary/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
