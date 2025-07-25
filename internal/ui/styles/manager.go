package styles

import (
	"sync"

	"github.com/charmbracelet/lipgloss"
)

// StyleManager is a global styles manager that provides themed styles
type StyleManager struct {
	themeManager *ThemeManager
	mu           sync.RWMutex
	listeners    []func()
}

var (
	globalStyleManager *StyleManager
	once               sync.Once
)

// GetStyleManager returns the global style manager instance
func GetStyleManager() *StyleManager {
	once.Do(func() {
		globalStyleManager = &StyleManager{
			themeManager: NewThemeManager(),
			listeners:    make([]func(), 0),
		}
	})
	return globalStyleManager
}

// GetTheme returns the current theme
func (sm *StyleManager) GetTheme() *Theme {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.themeManager.GetCurrentTheme()
}

// SetTheme changes the current theme and notifies listeners
func (sm *StyleManager) SetTheme(themeName string) error {
	sm.mu.Lock()
	err := sm.themeManager.SetTheme(themeName)
	sm.mu.Unlock()

	if err == nil {
		sm.notifyListeners()
	}

	return err
}

// ToggleTheme switches between light and dark themes
func (sm *StyleManager) ToggleTheme() {
	sm.mu.Lock()
	sm.themeManager.ToggleTheme()
	sm.mu.Unlock()

	sm.notifyListeners()
}

// AddThemeChangeListener adds a callback for theme changes
func (sm *StyleManager) AddThemeChangeListener(listener func()) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.listeners = append(sm.listeners, listener)
}

// notifyListeners notifies all registered listeners of theme change
func (sm *StyleManager) notifyListeners() {
	sm.mu.RLock()
	listeners := make([]func(), len(sm.listeners))
	copy(listeners, sm.listeners)
	sm.mu.RUnlock()

	for _, listener := range listeners {
		listener()
	}
}

// Component-specific style builders

// HeaderStyles returns styled components for headers
type HeaderStyles struct {
	Container     lipgloss.Style
	Title         lipgloss.Style
	Subtitle      lipgloss.Style
	ClusterInfo   lipgloss.Style
	NamespaceInfo lipgloss.Style
	Disconnected  lipgloss.Style
	Timestamp     lipgloss.Style
}

func (sm *StyleManager) GetHeaderStyles() HeaderStyles {
	theme := sm.GetTheme()

	return HeaderStyles{
		Container: CreateBaseStyle(theme).
			Align(lipgloss.Center),
		Title: CreatePrimaryStyle(theme).
			Bold(true),
		Subtitle: CreateSecondaryStyle(theme),
		ClusterInfo: CreateStatusStyle(theme, "success").
			Bold(true),
		NamespaceInfo: CreateSecondaryStyle(theme),
		Disconnected:  CreateStatusStyle(theme, "error"),
		Timestamp:     CreateMutedStyle(theme),
	}
}

// TabStyles returns styled components for tabs
type TabStyles struct {
	Container    lipgloss.Style
	ActiveTab    lipgloss.Style
	InactiveTab  lipgloss.Style
	TabSeparator lipgloss.Style
}

func (sm *StyleManager) GetTabStyles() TabStyles {
	theme := sm.GetTheme()

	return TabStyles{
		Container: CreateBaseStyle(theme).
			Align(lipgloss.Center),
		ActiveTab: lipgloss.NewStyle().
			Foreground(theme.AccentForeground).
			Background(theme.Primary).
			Padding(0, 1).
			Bold(true),
		InactiveTab: CreateMutedStyle(theme).
			Padding(0, 1),
		TabSeparator: CreateMutedStyle(theme),
	}
}

// PanelStyles returns styled components for panels
type PanelStyles struct {
	Container     lipgloss.Style
	Border        lipgloss.Style
	FocusedBorder lipgloss.Style
	Title         lipgloss.Style
	Content       lipgloss.Style
}

func (sm *StyleManager) GetPanelStyles(focused bool) PanelStyles {
	theme := sm.GetTheme()

	borderStyle := CreateBorderStyle(theme, focused)

	return PanelStyles{
		Container:     CreateBaseStyle(theme),
		Border:        borderStyle,
		FocusedBorder: CreateBorderStyle(theme, true),
		Title: CreatePrimaryStyle(theme).
			Align(lipgloss.Center),
		Content: CreateBaseStyle(theme).
			Padding(1),
	}
}

// StatusBarStyles returns styled components for status bar
type StatusBarStyles struct {
	Container     lipgloss.Style
	StatusText    lipgloss.Style
	KeyHint       lipgloss.Style
	ModeIndicator lipgloss.Style
}

func (sm *StyleManager) GetStatusBarStyles() StatusBarStyles {
	theme := sm.GetTheme()

	return StatusBarStyles{
		Container: lipgloss.NewStyle().
			Background(theme.Background).
			Foreground(theme.MutedForeground),
		StatusText:    CreateMutedStyle(theme),
		KeyHint:       CreateMutedStyle(theme),
		ModeIndicator: CreatePrimaryStyle(theme),
	}
}

// LogStyles returns styled components for log entries
type LogStyles struct {
	Container      lipgloss.Style
	TimestampStyle lipgloss.Style
	InfoStyle      lipgloss.Style
	WarnStyle      lipgloss.Style
	ErrorStyle     lipgloss.Style
	DebugStyle     lipgloss.Style
}

func (sm *StyleManager) GetLogStyles() LogStyles {
	theme := sm.GetTheme()

	return LogStyles{
		Container:      CreateBaseStyle(theme),
		TimestampStyle: CreateMutedStyle(theme),
		InfoStyle:      CreateStatusStyle(theme, "info"),
		WarnStyle:      CreateStatusStyle(theme, "warning"),
		ErrorStyle:     CreateStatusStyle(theme, "error"),
		DebugStyle:     CreateMutedStyle(theme),
	}
}

// ListStyles returns styled components for lists
type ListStyles struct {
	Container    lipgloss.Style
	Item         lipgloss.Style
	SelectedItem lipgloss.Style
	Header       lipgloss.Style
}

func (sm *StyleManager) GetListStyles() ListStyles {
	theme := sm.GetTheme()

	return ListStyles{
		Container:    CreateBaseStyle(theme),
		Item:         CreateBaseStyle(theme),
		SelectedItem: CreateSelectedStyle(theme),
		Header: CreatePrimaryStyle(theme).
			Underline(true),
	}
}

// DialogStyles returns styled components for dialogs/modals
type DialogStyles struct {
	Container    lipgloss.Style
	Title        lipgloss.Style
	Content      lipgloss.Style
	Button       lipgloss.Style
	ActiveButton lipgloss.Style
}

func (sm *StyleManager) GetDialogStyles() DialogStyles {
	theme := sm.GetTheme()

	return DialogStyles{
		Container: CreateBaseStyle(theme).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.Border).
			Padding(1, 2),
		Title: CreatePrimaryStyle(theme).
			Bold(true).
			MarginBottom(1),
		Content: CreateBaseStyle(theme).
			MarginBottom(1),
		Button: CreateSecondaryStyle(theme).
			Padding(0, 1).
			Border(lipgloss.NormalBorder()).
			BorderForeground(theme.Border),
		ActiveButton: lipgloss.NewStyle().
			Foreground(theme.AccentForeground).
			Background(theme.Primary).
			Padding(0, 1).
			Border(lipgloss.NormalBorder()).
			BorderForeground(theme.Primary),
	}
}
