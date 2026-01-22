package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("claude-insights-agent v0.1.0")
	if len(os.Args) < 2 {
		fmt.Println("Usage: claude-insights-agent <command>")
		fmt.Println("Commands: init, run, status, version")
		os.Exit(1)
	}
}
