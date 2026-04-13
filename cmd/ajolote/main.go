package main

import (
	"github.com/ajolote-ai/ajolote/internal/cli"
)

// version is set at build time via -ldflags "-X main.version=..."
var version = "dev"

func main() {
	cli.Execute(version)
}
