package main

import (
	"os"

	"spacereclaim/internal/cli"
)

func main() {
	os.Exit(cli.Execute())
}
