package ui

import (
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/katyella/lazyoc/internal/ui/messages"
	"github.com/katyella/lazyoc/internal/ui/models"
	"github.com/katyella/lazyoc/internal/ui/navigation"
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
	
	updatedTUI, ok := model.(*TUI)
	if !ok {
		t.Fatal("Model is not a *TUI")
	}
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
	
	updatedTUI, ok := model.(*TUI)
	if !ok {
		t.Fatal("Model is not a *TUI")
	}
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
	
	updatedTUI, ok := model.(*TUI)
	if !ok {
		t.Fatal("Model is not a *TUI")
	}
	if updatedTUI.State != models.StateHelp {
		t.Errorf("Expected state to be StateHelp after '?', got %v", updatedTUI.State)
	}
}

func TestTUIUpdate_KeyInput_TabNavigation(t *testing.T) {
	tui := NewTUI("0.1.0-test", false)
	
	// Initialize TUI with window size first and complete initialization
	windowMsg := tea.WindowSizeMsg{Width: 80, Height: 24}
	tui.Update(windowMsg)
	
	// Send init message to complete TUI setup
	initMsg := messages.InitMsg{}
	tui.Update(initMsg)
	
	tui.State = models.StateMain
	tui.ActiveTab = models.TabPods
	
	// Verify that NextTab method works directly
	initialTab := tui.ActiveTab
	tui.NextTab()
	if tui.ActiveTab == initialTab {
		t.Fatal("NextTab() method is not working directly")
	}
	
	// Reset for key test
	tui.ActiveTab = models.TabPods
	
	// Ensure navigation controller is properly set up
	if tui.navController == nil {
		t.Fatal("Navigation controller is nil")
	}
	
	// Test that the navigation controller recognizes the L key
	registry := tui.navController.GetRegistry()
	if registry == nil {
		t.Fatal("Keybinding registry is nil")
	}
	
	action, exists := registry.GetAction("L")
	if !exists {
		t.Fatal("L key not found in keybinding registry")
	}
	
	if action != navigation.ActionNextTab {
		t.Fatalf("Expected L key to map to ActionNextTab, got %v", action)
	}
	
	// Test L key (next tab) - simulate the full message flow
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'L'}}
	model, cmd := tui.Update(msg)
	
	// In BubbleTea, the command would be executed automatically
	// In the test, we need to manually execute the command and process the resulting message
	updatedTUI, ok := model.(*TUI)
	if !ok {
		t.Fatal("Model is not a *TUI")
	}
	
	if cmd != nil {
		// Execute the command to get the message
		followupMsg := cmd()
		if followupMsg != nil {
			// Check if it's a BatchMsg (from tea.Batch)
			if batchMsg, ok := followupMsg.(tea.BatchMsg); ok {
				// Execute only the first command to avoid multiple navigations
				if len(batchMsg) > 0 {
					msg := batchMsg[0]()
					if msg != nil {
						updatedTUI.Update(msg)
					}
				}
			} else {
				// Single message
				updatedTUI.Update(followupMsg)
			}
		}
	}
	
	if updatedTUI.ActiveTab != models.TabServices {
		t.Errorf("Expected tab to be TabServices (1) after L key, got %v (%d)", updatedTUI.GetTabName(updatedTUI.ActiveTab), updatedTUI.ActiveTab)
	}
}

func TestTUIView_Loading(t *testing.T) {
	tui := NewTUI("0.1.0-test", false)
	tui.SetDimensions(80, 24)
	tui.State = models.StateLoading
	tui.isReady = true // Mark as ready to allow rendering
	
	view := tui.View()
	if view == "" {
		t.Error("View should return content for loading state")
	}
	
	// Check for either "Loading" or "LazyOC" which are both expected in loading view
	if !containsString(view, "Loading") && !containsString(view, "LazyOC") {
		t.Errorf("Loading view should contain 'Loading' or 'LazyOC' text, got: %s", view)
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
	tui.isReady = true // Mark as ready to allow rendering
	
	view := tui.View()
	if view == "" {
		t.Error("View should return content for help state")
	}
	
	// Help view should contain either "Help" or "LazyOC Help" or navigation-related text
	if !containsString(view, "Help") && !containsString(view, "LazyOC") && !containsString(view, "Navigation") {
		t.Errorf("Help view should contain help-related text, got: %s", view)
	}
}

func TestTUIView_Error(t *testing.T) {
	tui := NewTUI("0.1.0-test", false)
	tui.SetDimensions(80, 24)
	tui.isReady = true // Mark as ready to allow rendering
	tui.SetError(errors.New("test error"))
	
	view := tui.View()
	if view == "" {
		t.Error("View should return content for error state")
	}
	
	// Error view should contain either the error message or "Error" text
	if !containsString(view, "test error") && !containsString(view, "Error") {
		t.Errorf("Error view should contain error-related text, got: %s", view)
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