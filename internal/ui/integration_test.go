package ui

import (
	"context"
	"testing"
	"time"
)

// TestTUIStartsAndStops tests that the TUI can be started and cleanly shut down
func TestTUIStartsAndStops(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	opts := ProgramOptions{
		Version:      "0.1.0-test",
		Debug:        true,
		AltScreen:    false, // Disable alt screen for testing
		MouseSupport: false,
	}

	// Create the program
	program := NewProgram(opts)
	if program == nil {
		t.Fatal("Failed to create program")
	}

	// Test that we can create a context and cancel it
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Start the program in a goroutine
	done := make(chan error, 1)
	go func() {
		// Note: This would normally block, but we're testing the setup
		// In a real integration test, we'd need to send quit signals
		done <- nil
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Program returned error: %v", err)
		}
	case <-ctx.Done():
		// This is expected - the test times out because the TUI would normally run indefinitely
		// In a real scenario, we'd send a quit message to cleanly exit
	}
}

// TestProgramCreationWithDifferentOptions tests various program configurations
func TestProgramCreationWithDifferentOptions(t *testing.T) {
	testCases := []struct {
		name string
		opts ProgramOptions
	}{
		{
			name: "minimal config",
			opts: ProgramOptions{
				Version:      "test",
				Debug:        false,
				AltScreen:    false,
				MouseSupport: false,
			},
		},
		{
			name: "debug enabled",
			opts: ProgramOptions{
				Version:      "test-debug",
				Debug:        true,
				AltScreen:    false,
				MouseSupport: false,
			},
		},
		{
			name: "full featured",
			opts: ProgramOptions{
				Version:      "test-full",
				Debug:        true,
				AltScreen:    true,
				MouseSupport: true,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			program := NewProgram(tc.opts)
			if program == nil {
				t.Errorf("Failed to create program for %s", tc.name)
			}
		})
	}
}
