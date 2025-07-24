// Package example demonstrates the modular TUI architecture
package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/katyella/lazyoc/internal/tui"
)

func main() {
	// Create the application
	app := tui.NewApp("1.0.0", false)

	// Set kubeconfig if provided
	if len(os.Args) > 1 {
		if err := app.SetKubeconfig(os.Args[1]); err != nil {
			log.Fatal(err)
		}
	}

	// Create the Bubble Tea program
	p := tea.NewProgram(app, tea.WithAltScreen())

	// Run the program
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}

// This example demonstrates:
// 1. Creating a modular TUI application
// 2. Component-based architecture with clear separation
// 3. State management through a centralized manager
// 4. Layout system that handles responsive design
// 5. Reusable components (header, tabs, panels, status bar)
//
// The architecture allows for:
// - Easy testing of individual components
// - Clear separation of concerns
// - Parallel development by multiple team members
// - Incremental refactoring from the monolithic design