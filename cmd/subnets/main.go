package main

import (
	"os"

	"github.com/jonhadfield/subnets/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
