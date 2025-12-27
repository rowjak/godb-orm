package main

import (
	"os"

	"github.com/rowjak/godb-orm/cmd"
)

func main() {
	// Dual-mode entry point:
	// - If arguments are provided, run in CLI mode
	// - If no arguments, launch GUI mode (Wails)

	if len(os.Args) > 1 {
		// CLI Mode: User provided command-line arguments
		cmd.Execute()
	} else {
		// GUI Mode: Launch Wails application
		StartGUI()
	}
}
