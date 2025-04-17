package tui

import (
	"github.com/Lunaris-Project/lunaris-installer/pkg/tui/ui"
	"github.com/charmbracelet/lipgloss"
)

// Colors - Tokyo Night theme (imported from ui package)
var (
	primaryColor    = ui.PrimaryColor
	secondaryColor  = ui.SecondaryColor
	successColor    = ui.SuccessColor
	warningColor    = ui.WarningColor
	errorColor      = ui.ErrorColor
	textColor       = ui.TextColor
	dimmedColor     = ui.DimmedColor
	accentColor     = ui.AccentColor
	backgroundColor = ui.BackgroundColor
)

// Styles
var (
	// Container for entire pages
	PageContainer = lipgloss.NewStyle().
			Align(lipgloss.Center).
			AlignVertical(lipgloss.Center)

	// Content box for sections
	ContentBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(1, 2).
			Align(lipgloss.Center)
	// Base text style
	BaseStyle = lipgloss.NewStyle().
			Foreground(textColor)

	// Title style
	TitleStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true).
			Underline(true).
			Padding(1, 0, 0, 0)

	// Subtitle style
	SubtitleStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Italic(true).
			Padding(0, 0, 1, 0)

	// Box style
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(accentColor).
			Padding(1, 2)

	// Button style
	ButtonStyle = lipgloss.NewStyle().
			Foreground(backgroundColor).
			Background(primaryColor).
			Bold(true).
			Padding(0, 3).
			Margin(1, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(accentColor)

	// Selection style
	SelectionStyle = lipgloss.NewStyle().
			Foreground(accentColor).
			Bold(true)

	// Highlight style
	HighlightStyle = lipgloss.NewStyle().
			Background(primaryColor).
			Foreground(backgroundColor).
			Bold(true).
			Padding(0, 1)

	// Info style
	InfoStyle = lipgloss.NewStyle().
			Foreground(textColor)

	// Warning style
	WarningStyle = lipgloss.NewStyle().
			Foreground(warningColor).
			Bold(true)

	// Error style
	ErrorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true)

	// Success style
	SuccessStyle = lipgloss.NewStyle().
			Foreground(successColor).
			Bold(true)

	// Dim style
	DimStyle = lipgloss.NewStyle().
			Foreground(dimmedColor)

	// Focused style for inputs
	FocusedStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(1, 2)

	// Unfocused style for inputs
	UnfocusedStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(dimmedColor).
			Padding(1, 2)
)

// RenderCheckbox renders a checkbox
func RenderCheckbox(checked bool) string {
	return ui.Checkbox(checked, "", false)
}

// RenderProgressBar renders a progress bar
func RenderProgressBar(width, percent int) string {
	return ui.ProgressBar(width, percent)
}
