package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/katyella/lazyoc/internal/logging"
)

// ProgramOptions holds configuration for the Bubble Tea program
type ProgramOptions struct {
	Version             string
	Debug               bool
	AltScreen           bool
	MouseSupport        bool
	KubeConfig          string
	ShowFullClusterInfo bool
}

// DefaultProgramOptions returns sensible defaults for the TUI program
func DefaultProgramOptions() ProgramOptions {
	return ProgramOptions{
		Version:             "dev",
		Debug:               false,
		AltScreen:           true,  // Use alternate screen buffer
		MouseSupport:        true,  // Enable mouse support for scrolling
		ShowFullClusterInfo: false, // Obfuscate cluster info by default for security
	}
}

// NewProgram creates a new Bubble Tea program with the TUI model
func NewProgram(opts ProgramOptions) *tea.Program {
	// Create the simplified TUI model
	tui := NewTUI(opts.Version, opts.Debug, opts.ShowFullClusterInfo)

	// Set kubeconfig if provided
	if opts.KubeConfig != "" {
		tui.KubeconfigPath = opts.KubeConfig
	}

	// Configure program options
	var programOpts []tea.ProgramOption

	if opts.AltScreen {
		programOpts = append(programOpts, tea.WithAltScreen())
	}

	if opts.MouseSupport {
		// Use all motion for better trackpad support
		programOpts = append(programOpts, tea.WithMouseAllMotion())
	}

	// Add input handling (using default stdin, no need to specify nil)
	// programOpts = append(programOpts, tea.WithInput(nil)) // Use stdin

	logging.Info(tui.Logger, "Creating Bubble Tea program with options: AltScreen=%v, Mouse=%v",
		opts.AltScreen, opts.MouseSupport)

	// Create the program
	program := tea.NewProgram(tui, programOpts...)

	// Set program reference in TUI for goroutine message sending
	tui.program = program

	return program
}

// RunTUI creates and runs the TUI with the given options
func RunTUI(opts ProgramOptions) error {
	program := NewProgram(opts)

	// Start the program
	model, err := program.Run()
	if err != nil {
		return err
	}

	// Log final state if debug is enabled
	if tui, ok := model.(*TUI); ok && tui.Debug {
		logging.Info(tui.Logger, "TUI program exited successfully")
	}

	return nil
}
