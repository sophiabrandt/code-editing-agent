package main

import (
	"os"

	"github.com/sophiabrandt/code-editing-agent/agent"
)

var (
	version = "dev"
	commit  = "none"
)

func main() {
	os.Exit(agent.CLI(os.Args[1:], version, commit))
}
