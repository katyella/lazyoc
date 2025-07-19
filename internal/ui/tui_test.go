package ui

import (
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattwojtowicz/lazyoc/internal/ui/messages"
	"github.com/mattwojtowicz/lazyoc/internal/ui/models"
)

func TestNewTUI(t *testing.T) {
	tui := NewTUI("0.1.0-test", false)
	
	if tui.Version != "0.1.0-test" {
		t.Errorf("Expected version 0.1.0-test, got %s", tui.Version)
	}
	
	if tui.State != models.StateLoading {
		t.Errorf("Expected initial state to be StateLoading, got %v", tui.State)
	}
	
	if tui.ActiveTab != models.TabPods {
		t.Errorf("Expected initial tab to be TabPods, got %v", tui.ActiveTab)
	}
}

func TestTUIInit(t *testing.T) {
	tui := NewTUI("0.1.0-test", false)
	
	cmd := tui.Init()
	if cmd == nil {
		t.Error("Init() should return a command")
	}
}

func TestTUIUpdate_WindowSize(t *testing.T) {
	tui := NewTUI("0.1.0-test", false)
	
	msg := tea.WindowSizeMsg{Width: 80, Height: 24}
	model, cmd := tui.Update(msg)
	
	updatedTUI := model.(*TUI)
	if updatedTUI.Width != 80 || updatedTUI.Height != 24 {
		t.Errorf("Expected dimensions 80x24, got %dx%d", updatedTUI.Width, updatedTUI.Height)
	}
	
	if cmd != nil {
		t.Error("WindowSizeMsg should not return a command")
	}
}

func TestTUIUpdate_Init(t *testing.T) {
	tui := NewTUI("0.1.0-test", false)
	
	msg := messages.InitMsg{}
	model, _ := tui.Update(msg)
	
	updatedTUI := model.(*TUI)
	if updatedTUI.State != models.StateMain {
		t.Errorf("Expected state to be StateMain after init, got %v", updatedTUI.State)
	}
	
	if updatedTUI.Loading {
		t.Error("Expected loading to be false after init")
	}
}

func TestTUIUpdate_KeyInput_Quit(t *testing.T) {
	tui := NewTUI("0.1.0-test", false)
	tui.State = models.StateMain
	
	// Test 'q' key
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	_, cmd := tui.Update(msg)
	
	if cmd == nil {
		t.Error("Quit key should return tea.Quit command")
	}
}

func TestTUIUpdate_KeyInput_Help(t *testing.T) {
	tui := NewTUI("0.1.0-test", false)
	tui.State = models.StateMain
	
	// Test '?' key
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}
	model, _ := tui.Update(msg)
	
	updatedTUI := model.(*TUI)
	if updatedTUI.State != models.StateHelp {
		t.Errorf("Expected state to be StateHelp after '?', got %v", updatedTUI.State)
	}
}

func TestTUIUpdate_KeyInput_TabNavigation(t *testing.T) {
	tui := NewTUI("0.1.0-test", false)
	tui.State = models.StateMain
	tui.ActiveTab = models.TabPods
	
	// Test tab key (next tab)
	msg := tea.KeyMsg{Type: tea.KeyTab}
	model, _ := tui.Update(msg)
	
	updatedTUI := model.(*TUI)
	if updatedTUI.ActiveTab != models.TabServices {
		t.Errorf("Expected tab to be TabServices after tab key, got %v", updatedTUI.ActiveTab)
	}
}

func TestTUIView_Loading(t *testing.T) {
	tui := NewTUI("0.1.0-test", false)
	tui.SetDimensions(80, 24)
	tui.State = models.StateLoading
	
	view := tui.View()
	if view == "" {
		t.Error("View should return content for loading state")
	}
	
	if !containsString(view, "Loading") {
		t.Error("Loading view should contain 'Loading' text")
	}
}

func TestTUIView_Main(t *testing.T) {
	tui := NewTUI("0.1.0-test", false)
	tui.SetDimensions(80, 24)
	tui.State = models.StateMain
	
	view := tui.View()
	if view == "" {
		t.Error("View should return content for main state")
	}
	
	if !containsString(view, "LazyOC") {
		t.Error("Main view should contain 'LazyOC' text")
	}
}

func TestTUIView_Help(t *testing.T) {
	tui := NewTUI("0.1.0-test", false)
	tui.SetDimensions(80, 24)
	tui.State = models.StateHelp
	
	view := tui.View()
	if view == "" {
		t.Error("View should return content for help state")
	}
	
	if !containsString(view, "Help") {
		t.Error("Help view should contain 'Help' text")
	}
}

func TestTUIView_Error(t *testing.T) {
	tui := NewTUI("0.1.0-test", false)
	tui.SetDimensions(80, 24)
	tui.SetError(errors.New("test error"))
	
	view := tui.View()
	if view == "" {
		t.Error("View should return content for error state")
	}
	
	if !containsString(view, "test error") {
		t.Error("Error view should contain error message")
	}
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && 
		   func() bool {
			   for i := 0; i <= len(s)-len(substr); i++ {
				   if s[i:i+len(substr)] == substr {
					   return true
				   }
			   }
			   return false
		   }()
}