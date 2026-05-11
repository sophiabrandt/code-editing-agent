package main

import (
	"os"

	"github.com/sophiabrandt/code-editing-agent/app"
)

var (
	version = "dev"
	commit  = "none"
)

func main() {
	os.Exit(app.CLI(os.Args[1:], version, commit))
}
