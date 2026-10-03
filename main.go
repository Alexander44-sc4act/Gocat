package main

import (
	"fmt"
	"os"

	"github.com/realibrahimsql/Gocat/cmd"
)

// Build information (set by ldflags)
var (
	version   = "dev"
	buildTime = "unknown"
	gitCommit = "unknown"
	gitBranch = "unknown"
	builtBy   = "unknown"
)

func main() {
	cmd.SetBuildInfo(version, buildTime, gitCommit, gitBranch, builtBy)

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v", err)
		os.Exit(1)
	}
}
