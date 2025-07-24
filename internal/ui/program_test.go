package ui

import (
	"testing"
	
	tea "github.com/charmbracelet/bubbletea"
)

func TestDefaultProgramOptions(t *testing.T) {
	opts := DefaultProgramOptions()
	
	if opts.Version != "dev" {
		t.Errorf("Expected default version 'dev', got %s", opts.Version)
	}
	
	if opts.Debug != false {
		t.Errorf("Expected debug to be false by default, got %v", opts.Debug)
	}
	
	if opts.AltScreen != true {
		t.Errorf("Expected AltScreen to be true by default, got %v", opts.AltScreen)
	}
	
	if opts.MouseSupport != false {
		t.Errorf("Expected MouseSupport to be false by default, got %v", opts.MouseSupport)
	}
}

func TestNewProgram(t *testing.T) {
	opts := ProgramOptions{
		Version:      "0.1.0-test",
		Debug:        true,
		AltScreen:    true,
		MouseSupport: false,
	}
	
	program := NewProgram(opts)
	
	if program == nil {
		t.Fatal("NewProgram should return a non-nil program")
	}
	
	// Test that we can get the model
	// Note: We can't easily test the actual program execution in a unit test
	// without running the full TUI, but we can verify the program was created
}

func TestProgramOptionsConfiguration(t *testing.T) {
	testCases := []struct {
		name string
		opts ProgramOptions
	}{
		{
			name: "debug enabled",
			opts: ProgramOptions{
				Version:      "test",
				Debug:        true,
				AltScreen:    true,
				MouseSupport: false,
			},
		},
		{
			name: "alt screen disabled",
			opts: ProgramOptions{
				Version:      "test",
				Debug:        false,
				AltScreen:    false,
				MouseSupport: false,
			},
		},
		{
			name: "mouse support enabled",
			opts: ProgramOptions{
				Version:      "test",
				Debug:        false,
				AltScreen:    true,
				MouseSupport: true,
			},
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			program := NewProgram(tc.opts)
			if program == nil {
				t.Errorf("NewProgram returned nil for test case: %s", tc.name)
			}
		})
	}
}

// Test that the TUI model works with Bubble Tea program
func TestTUIModelIntegration(t *testing.T) {
	tui := NewTUI("0.1.0-test", false)
	
	// Verify it implements tea.Model
	var _ tea.Model = tui
	
	// Test the basic lifecycle
	cmd := tui.Init()
	if cmd == nil {
		t.Error("Init should return a command")
	}
	
	// Test a simple update
	model, _ := tui.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	updatedTUI, ok := model.(*TUI)
	if !ok {
		t.Fatal("Model is not a *TUI")
	}
	
	if updatedTUI.Width != 80 || updatedTUI.Height != 24 {
		t.Errorf("Expected dimensions 80x24, got %dx%d", updatedTUI.Width, updatedTUI.Height)
	}
	
	// Test view rendering
	view := updatedTUI.View()
	if view == "" {
		t.Error("View should return non-empty content")
	}
}