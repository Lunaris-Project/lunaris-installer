package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Colors
var (
	primaryColor   = lipgloss.Color("#7dcfff")
	secondaryColor = lipgloss.Color("#bb9af7")
	successColor   = lipgloss.Color("#9ece6a")
	warningColor   = lipgloss.Color("#e0af68")
	errorColor     = lipgloss.Color("#f7768e")
	textColor      = lipgloss.Color("#c0caf5")
	dimmedColor    = lipgloss.Color("#565f89")
)

// Styles
var (
	// Base text style
	BaseStyle = lipgloss.NewStyle().
			Foreground(textColor)

	// Title style
	TitleStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true).
			Padding(1, 0, 0, 0)

	// Subtitle style
	SubtitleStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Bold(true).
			Padding(0, 0, 1, 0)

	// Box style
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(1, 2)

	// Button style
	ButtonStyle = lipgloss.NewStyle().
			Foreground(textColor).
			Background(dimmedColor).
			Bold(true).
			Padding(0, 3)

	// Selection style
	SelectionStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true)

	// Highlight style
	HighlightStyle = lipgloss.NewStyle().
			Background(primaryColor).
			Foreground(lipgloss.Color("#1a1b26"))

	// Info style
	InfoStyle = lipgloss.NewStyle().
			Foreground(textColor)

	// Warning style
	WarningStyle = lipgloss.NewStyle().
			Foreground(warningColor)

	// Error style
	ErrorStyle = lipgloss.NewStyle().
			Foreground(errorColor)

	// Success style
	SuccessStyle = lipgloss.NewStyle().
			Foreground(successColor)

	// Dim style
	DimStyle = lipgloss.NewStyle().
			Foreground(dimmedColor)
)

// RenderCheckbox renders a checkbox
func RenderCheckbox(checked bool) string {
	if checked {
		return "[✓]"
	}
	return "[ ]"
}

// RenderProgressBar renders a progress bar
func RenderProgressBar(width, percent int) string {
	// Ensure percent is between 0 and 100
	if percent < 0 {
		percent = 0
	} else if percent > 100 {
		percent = 100
	}

	// Calculate the width of the filled portion
	filledWidth := (width * percent) / 100

	// Create the filled and empty portions
	filled := strings.Repeat("█", filledWidth)
	empty := strings.Repeat("░", width-filledWidth)

	// Combine and style
	bar := lipgloss.NewStyle().Foreground(primaryColor).Render(filled) +
		lipgloss.NewStyle().Foreground(dimmedColor).Render(empty)

	return bar
}
