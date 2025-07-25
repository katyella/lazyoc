package messages

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestMessagePriority(t *testing.T) {
	handler := NewMessageHandler()

	tests := []struct {
		name     string
		msg      tea.Msg
		expected MessagePriority
	}{
		{
			name:     "Quit message is critical",
			msg:      tea.QuitMsg{},
			expected: PriorityCritical,
		},
		{
			name:     "Error message is critical",
			msg:      ErrorMsg{},
			expected: PriorityCritical,
		},
		{
			name:     "Window size is high priority",
			msg:      tea.WindowSizeMsg{Width: 80, Height: 24},
			expected: PriorityHigh,
		},
		{
			name:     "Init message is high priority",
			msg:      InitMsg{},
			expected: PriorityHigh,
		},
		{
			name:     "Key message is normal priority",
			msg:      tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")},
			expected: PriorityNormal,
		},
		{
			name:     "Refresh message is low priority",
			msg:      RefreshMsg{},
			expected: PriorityLow,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			priority := handler.GetMessagePriority(tt.msg)
			if priority != tt.expected {
				t.Errorf("GetMessagePriority() = %v, want %v", priority, tt.expected)
			}
		})
	}
}

func TestMessageTracking(t *testing.T) {
	handler := NewMessageHandler()

	// Track a message multiple times
	for i := 0; i < 150; i++ {
		handler.TrackMessage(StatusMsg{Message: "test"})
	}

	// Should be marked as high frequency after 100 occurrences
	if !handler.IsHighFrequency(StatusMsg{}) {
		t.Error("Expected StatusMsg to be marked as high frequency")
	}

	// Other message types should not be high frequency
	if handler.IsHighFrequency(InitMsg{}) {
		t.Error("Expected InitMsg to not be marked as high frequency")
	}
}

func TestMessageBatching(t *testing.T) {
	handler := NewMessageHandler()

	// Critical messages should never be batched
	if handler.ShouldBatch(tea.QuitMsg{}) {
		t.Error("Critical messages should not be batched")
	}

	// Track status messages to make them high frequency
	for i := 0; i < 150; i++ {
		handler.TrackMessage(StatusMsg{Message: "test"})
	}

	// High frequency status messages should be batched
	if !handler.ShouldBatch(StatusMsg{Message: "test"}) {
		t.Error("High frequency status messages should be batched")
	}
}

func TestMessageDeduplication(t *testing.T) {
	handler := NewMessageHandler()

	// Add multiple similar messages
	handler.AddToBatch(StatusMsg{Message: "status1", Type: StatusInfo})
	handler.AddToBatch(StatusMsg{Message: "status2", Type: StatusInfo})
	handler.AddToBatch(StatusMsg{Message: "status3", Type: StatusInfo})
	handler.AddToBatch(LoadingMsg{Message: "loading1"})
	handler.AddToBatch(LoadingMsg{Message: "loading2"})

	// Process batch
	processed := handler.ProcessBatch()

	// Should have deduplicated to 2 messages (one StatusMsg, one LoadingMsg)
	if len(processed) != 2 {
		t.Errorf("Expected 2 deduplicated messages, got %d", len(processed))
	}

	// Verify the latest messages are kept
	foundStatus := false
	foundLoading := false

	for _, msg := range processed {
		switch m := msg.(type) {
		case StatusMsg:
			if m.Message != "status3" {
				t.Error("Expected latest status message to be kept")
			}
			foundStatus = true
		case LoadingMsg:
			if m.Message != "loading2" {
				t.Error("Expected latest loading message to be kept")
			}
			foundLoading = true
		}
	}

	if !foundStatus || !foundLoading {
		t.Error("Expected both status and loading messages in processed batch")
	}
}

func TestProcessWithHandlers(t *testing.T) {
	// Test handler chaining
	handler1Called := false
	handler2Called := false

	handler1 := func(msg tea.Msg) (bool, []tea.Cmd) {
		handler1Called = true
		// Don't handle the message, let it pass through
		return false, nil
	}

	handler2 := func(msg tea.Msg) (bool, []tea.Cmd) {
		handler2Called = true
		// Handle the message
		return true, []tea.Cmd{func() tea.Msg { return nil }}
	}

	cmds := ProcessWithHandlers(InitMsg{}, handler1, handler2)

	if !handler1Called {
		t.Error("Expected handler1 to be called")
	}

	if !handler2Called {
		t.Error("Expected handler2 to be called")
	}

	if len(cmds) != 1 {
		t.Error("Expected one command from handler2")
	}
}
