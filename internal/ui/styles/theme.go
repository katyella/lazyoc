package styles

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/charmbracelet/lipgloss"
)

// Theme represents a color theme for the UI
type Theme struct {
	Name string `json:"name"`
	
	// Basic colors
	Background lipgloss.Color `json:"background"`
	Foreground lipgloss.Color `json:"foreground"`
	
	// UI colors
	Primary   lipgloss.Color `json:"primary"`
	Secondary lipgloss.Color `json:"secondary"`
	Border    lipgloss.Color `json:"border"`
	
	// Status colors
	Success lipgloss.Color `json:"success"`
	Warning lipgloss.Color `json:"warning"`
	Error   lipgloss.Color `json:"error"`
	Info    lipgloss.Color `json:"info"`
	
	// Additional UI colors
	AccentForeground lipgloss.Color `json:"accent_foreground"`
	MutedForeground  lipgloss.Color `json:"muted_foreground"`
	SelectedBg       lipgloss.Color `json:"selected_bg"`
	FocusBorder      lipgloss.Color `json:"focus_border"`
}

// PredefinedThemes contains the built-in themes
var PredefinedThemes = map[string]*Theme{
	"dark": {
		Name:             "dark",
		Background:       lipgloss.Color("0"),   // Black
		Foreground:       lipgloss.Color("15"),  // White
		Primary:          lipgloss.Color("12"),  // Blue
		Secondary:        lipgloss.Color("14"),  // Cyan
		Border:           lipgloss.Color("8"),   // Gray
		Success:          lipgloss.Color("10"),  // Green
		Warning:          lipgloss.Color("11"),  // Yellow
		Error:            lipgloss.Color("9"),   // Red
		Info:             lipgloss.Color("12"),  // Blue
		AccentForeground: lipgloss.Color("15"),  // White
		MutedForeground:  lipgloss.Color("8"),   // Gray
		SelectedBg:       lipgloss.Color("8"),   // Gray
		FocusBorder:      lipgloss.Color("12"),  // Blue
	},
	"light": {
		Name:             "light",
		Background:       lipgloss.Color("15"),  // White
		Foreground:       lipgloss.Color("0"),   // Black
		Primary:          lipgloss.Color("4"),   // Dark Blue
		Secondary:        lipgloss.Color("6"),   // Dark Cyan
		Border:           lipgloss.Color("7"),   // Light Gray
		Success:          lipgloss.Color("2"),   // Dark Green
		Warning:          lipgloss.Color("3"),   // Dark Yellow
		Error:            lipgloss.Color("1"),   // Dark Red
		Info:             lipgloss.Color("4"),   // Dark Blue
		AccentForeground: lipgloss.Color("0"),   // Black
		MutedForeground:  lipgloss.Color("8"),   // Gray
		SelectedBg:       lipgloss.Color("7"),   // Light Gray
		FocusBorder:      lipgloss.Color("4"),   // Dark Blue
	},
}

// ThemeManager manages theme configuration and persistence
type ThemeManager struct {
	currentTheme *Theme
	configPath   string
}

// ThemeConfig represents the persisted theme configuration
type ThemeConfig struct {
	SelectedTheme string `json:"selected_theme"`
}

// NewThemeManager creates a new theme manager instance
func NewThemeManager() *ThemeManager {
	configDir := filepath.Join(os.Getenv("HOME"), ".lazyoc")
	configPath := filepath.Join(configDir, "config.json")
	
	tm := &ThemeManager{
		currentTheme: PredefinedThemes["dark"], // Default to dark theme
		configPath:   configPath,
	}
	
	// Ensure config directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		// Silently continue if config directory creation fails
	}
	
	// Load saved theme preference
	tm.loadThemePreference()
	
	return tm
}

// GetCurrentTheme returns the currently active theme
func (tm *ThemeManager) GetCurrentTheme() *Theme {
	return tm.currentTheme
}

// SetTheme switches to the specified theme
func (tm *ThemeManager) SetTheme(themeName string) error {
	theme, exists := PredefinedThemes[themeName]
	if !exists {
		theme = PredefinedThemes["dark"] // fallback
	}
	
	tm.currentTheme = theme
	
	// Save theme preference
	if err := tm.saveThemePreference(); err != nil {
		return err
	}
	
	return nil
}

// ToggleTheme switches between light and dark themes
func (tm *ThemeManager) ToggleTheme() {
	if tm.currentTheme.Name == "dark" {
		tm.SetTheme("light")
	} else {
		tm.SetTheme("dark")
	}
}

// loadThemePreference loads the saved theme preference from disk
func (tm *ThemeManager) loadThemePreference() {
	data, err := os.ReadFile(tm.configPath)
	if err != nil {
		if !os.IsNotExist(err) {
			// Silently ignore read errors
		}
		return
	}
	
	var config ThemeConfig
	if err := json.Unmarshal(data, &config); err != nil {
		// Silently ignore parse errors
		return
	}
	
	if config.SelectedTheme != "" {
		tm.SetTheme(config.SelectedTheme)
	}
}

// saveThemePreference saves the current theme preference to disk
func (tm *ThemeManager) saveThemePreference() error {
	config := ThemeConfig{
		SelectedTheme: tm.currentTheme.Name,
	}
	
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(tm.configPath, data, 0644)
}

// Style helper functions for creating themed styles

// CreateBaseStyle creates a base style with theme colors
func CreateBaseStyle(theme *Theme) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Foreground).
		Background(theme.Background)
}

// CreateBorderStyle creates a bordered style with theme colors
func CreateBorderStyle(theme *Theme, focused bool) lipgloss.Style {
	borderColor := theme.Border
	if focused {
		borderColor = theme.FocusBorder
	}
	
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor)
}

// CreatePrimaryStyle creates a style with primary theme colors
func CreatePrimaryStyle(theme *Theme) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Primary).
		Bold(true)
}

// CreateSecondaryStyle creates a style with secondary theme colors
func CreateSecondaryStyle(theme *Theme) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Secondary)
}

// CreateMutedStyle creates a style with muted colors
func CreateMutedStyle(theme *Theme) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.MutedForeground)
}

// CreateStatusStyle creates a style based on status type
func CreateStatusStyle(theme *Theme, statusType string) lipgloss.Style {
	var color lipgloss.Color
	
	switch statusType {
	case "success":
		color = theme.Success
	case "warning":
		color = theme.Warning
	case "error":
		color = theme.Error
	case "info":
		color = theme.Info
	default:
		color = theme.Foreground
	}
	
	return lipgloss.NewStyle().Foreground(color)
}

// CreateSelectedStyle creates a style for selected items
func CreateSelectedStyle(theme *Theme) lipgloss.Style {
	return lipgloss.NewStyle().
		Background(theme.SelectedBg).
		Foreground(theme.AccentForeground)
}