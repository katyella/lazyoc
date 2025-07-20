package ui

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SimpleTUI is a minimal working TUI for debugging
type SimpleTUI struct {
	width   int
	height  int
	version string
	debug   bool
}

func NewSimpleTUI(version string, debug bool) *SimpleTUI {
	return &SimpleTUI{
		version: version,
		debug:   debug,
	}
}

func (s *SimpleTUI) Init() tea.Cmd {
	return nil
}

func (s *SimpleTUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return s, tea.Quit
		}
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
	}
	return s, nil
}

func (s *SimpleTUI) View() string {
	if s.width == 0 || s.height == 0 {
		return "Initializing..."
	}
	
	style := lipgloss.NewStyle().
		Width(s.width).
		Height(s.height).
		Align(lipgloss.Center, lipgloss.Center)
	
	content := fmt.Sprintf("🚀 LazyOC v%s\n\nSimple Mode - Press 'q' to quit\n\nTerminal: %dx%d", 
		s.version, s.width, s.height)
	
	return style.Render(content)
}