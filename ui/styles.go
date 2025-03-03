package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Colors
	primaryColor   = lipgloss.Color("#7D56F4")
	secondaryColor = lipgloss.Color("#FC7A57")
	successColor   = lipgloss.Color("#73F59F")
	errorColor     = lipgloss.Color("#F25D94")
	warningColor   = lipgloss.Color("#F2C14E")
	infoColor      = lipgloss.Color("#4D9DE0")
	textColor      = lipgloss.Color("#FFFFFF")
	dimTextColor   = lipgloss.Color("#AAAAAA")
	bgColor        = lipgloss.Color("#1A1A1A")

	// Base styles
	BaseStyle = lipgloss.NewStyle().
			Foreground(textColor).
			Background(bgColor).
			PaddingLeft(2).
			PaddingRight(2)

	TitleStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Background(bgColor).
			Bold(true).
			PaddingLeft(2).
			PaddingRight(2).
			MarginBottom(1)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Background(bgColor).
			Bold(true).
			PaddingLeft(2).
			PaddingRight(2).
			MarginBottom(1)

	InfoStyle = lipgloss.NewStyle().
			Foreground(infoColor).
			Background(bgColor).
			PaddingLeft(2).
			PaddingRight(2)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(successColor).
			Background(bgColor).
			PaddingLeft(2).
			PaddingRight(2)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Background(bgColor).
			PaddingLeft(2).
			PaddingRight(2)

	WarningStyle = lipgloss.NewStyle().
			Foreground(warningColor).
			Background(bgColor).
			PaddingLeft(2).
			PaddingRight(2)

	DimStyle = lipgloss.NewStyle().
			Foreground(dimTextColor).
			Background(bgColor).
			PaddingLeft(2).
			PaddingRight(2)

	HighlightStyle = lipgloss.NewStyle().
			Foreground(bgColor).
			Background(primaryColor).
			Bold(true).
			PaddingLeft(1).
			PaddingRight(1)

	SelectionStyle = lipgloss.NewStyle().
			Foreground(bgColor).
			Background(secondaryColor).
			Bold(true).
			PaddingLeft(1).
			PaddingRight(1)

	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(1).
			MarginTop(1).
			MarginBottom(1)

	FocusedBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(secondaryColor).
			Padding(1).
			MarginTop(1).
			MarginBottom(1)

	ButtonStyle = lipgloss.NewStyle().
			Foreground(bgColor).
			Background(primaryColor).
			Bold(true).
			Padding(0, 3).
			MarginRight(1)

	ActiveButtonStyle = lipgloss.NewStyle().
				Foreground(bgColor).
				Background(secondaryColor).
				Bold(true).
				Padding(0, 3).
				MarginRight(1)

	DisabledButtonStyle = lipgloss.NewStyle().
				Foreground(bgColor).
				Background(dimTextColor).
				Bold(true).
				Padding(0, 3).
				MarginRight(1)

	CheckboxCheckedStyle = lipgloss.NewStyle().
				Foreground(successColor).
				Background(bgColor).
				PaddingRight(1)

	CheckboxUncheckedStyle = lipgloss.NewStyle().
				Foreground(dimTextColor).
				Background(bgColor).
				PaddingRight(1)

	ProgressBarFilledStyle = lipgloss.NewStyle().
				Foreground(bgColor).
				Background(primaryColor)

	ProgressBarEmptyStyle = lipgloss.NewStyle().
				Foreground(bgColor).
				Background(dimTextColor)

	StatusBarStyle = lipgloss.NewStyle().
			Foreground(textColor).
			Background(primaryColor).
			Bold(true).
			Width(100).
			Align(lipgloss.Center)
)

// RenderCheckbox renders a checkbox
func RenderCheckbox(checked bool, label string) string {
	checkbox := "[ ]"
	style := CheckboxUncheckedStyle

	if checked {
		checkbox = "[✓]"
		style = CheckboxCheckedStyle
	}

	return style.Render(checkbox) + " " + label
}

// RenderProgressBar renders a progress bar
func RenderProgressBar(width, percent int) string {
	// Ensure percent is within valid range
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}

	// Calculate filled and empty portions
	filled := int(float64(width) * float64(percent) / 100.0)

	// Ensure filled is not negative
	if filled < 0 {
		filled = 0
	}

	// Ensure filled is not greater than width
	if filled > width {
		filled = width
	}

	empty := width - filled

	filledBar := ProgressBarFilledStyle.Render(strings.Repeat("█", filled))
	emptyBar := ProgressBarEmptyStyle.Render(strings.Repeat("█", empty))

	return filledBar + emptyBar
}

// RenderStatusBar renders a status bar
func RenderStatusBar(status string) string {
	return StatusBarStyle.Render(status)
}
