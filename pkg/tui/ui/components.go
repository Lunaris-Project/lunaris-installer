package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Colors - Tokyo Night theme
var (
	PrimaryColor    = lipgloss.Color("#7dcfff") // Light blue
	SecondaryColor  = lipgloss.Color("#bb9af7") // Purple
	SuccessColor    = lipgloss.Color("#9ece6a") // Green
	WarningColor    = lipgloss.Color("#e0af68") // Yellow/Orange
	ErrorColor      = lipgloss.Color("#f7768e") // Red/Pink
	TextColor       = lipgloss.Color("#c0caf5") // Light text
	DimmedColor     = lipgloss.Color("#565f89") // Dimmed text
	AccentColor     = lipgloss.Color("#2ac3de") // Cyan
	BackgroundColor = lipgloss.Color("#1a1b26") // Dark background
)

// Container creates a container with the given content
func Container(content string, width, height int) string {
	return lipgloss.NewStyle().
		Width(width).
		Height(height).
		Align(lipgloss.Center).
		AlignVertical(lipgloss.Center).
		Render(content)
}

// Box creates a box with the given content
func Box(content string, width int, title string) string {
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(PrimaryColor).
		Padding(1, 2).
		Width(width)

	if title != "" {
		boxStyle = boxStyle.BorderTop(true).
			BorderLeft(true).
			BorderRight(true).
			BorderBottom(true)

		// Create a title for the box
		titleStyle := lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Bold(true).
			Padding(0, 1)

		renderedTitle := titleStyle.Render(title)

		// Calculate the width of the title
		titleWidth := lipgloss.Width(renderedTitle)

		// Calculate the width of the box
		boxWidth := width - 4 // Subtract padding and borders

		// Calculate the left padding
		leftPadding := (boxWidth - titleWidth) / 2
		if leftPadding < 0 {
			leftPadding = 0
		}

		// Create the title line (unused for now, but could be used for custom title rendering)
		_ = strings.Repeat("─", leftPadding) + " " + renderedTitle + " " + strings.Repeat("─", boxWidth-titleWidth-leftPadding-2)

		// Render the box with the title
		return boxStyle.Border(lipgloss.RoundedBorder()).
			BorderForeground(PrimaryColor).
			BorderTop(false).
			Render(content)
	}

	return boxStyle.Render(content)
}

// Title creates a title
func Title(content string, width int) string {
	return lipgloss.NewStyle().
		Foreground(PrimaryColor).
		Bold(true).
		Width(width).
		Align(lipgloss.Center).
		Render(content)
}

// Subtitle creates a subtitle
func Subtitle(content string, width int) string {
	return lipgloss.NewStyle().
		Foreground(SecondaryColor).
		Italic(true).
		Width(width).
		Align(lipgloss.Center).
		Render(content)
}

// Button creates a button
func Button(content string, selected bool) string {
	style := lipgloss.NewStyle().
		Padding(0, 3).
		Margin(1, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(AccentColor)

	if selected {
		style = style.
			Foreground(BackgroundColor).
			Background(PrimaryColor).
			Bold(true)
	} else {
		style = style.
			Foreground(TextColor)
	}

	return style.Render(content)
}

// Option creates an option
func Option(content string, selected bool) string {
	if selected {
		return lipgloss.NewStyle().
			Foreground(AccentColor).
			Bold(true).
			Render("> " + content)
	}
	return lipgloss.NewStyle().
		Foreground(TextColor).
		Render("  " + content)
}

// Checkbox creates a checkbox
func Checkbox(checked bool, label string, selected bool) string {
	var checkbox string
	if checked {
		checkbox = lipgloss.NewStyle().
			Foreground(SuccessColor).
			Bold(true).
			Render("[✓]")
	} else {
		checkbox = lipgloss.NewStyle().
			Foreground(DimmedColor).
			Render("[ ]")
	}

	labelStyle := lipgloss.NewStyle()
	if selected {
		labelStyle = labelStyle.
			Foreground(AccentColor).
			Bold(true)
	} else {
		labelStyle = labelStyle.
			Foreground(TextColor)
	}

	return lipgloss.JoinHorizontal(lipgloss.Left, checkbox, " ", labelStyle.Render(label))
}

// ProgressBar creates a progress bar
func ProgressBar(width, percent int) string {
	// Ensure percent is between 0 and 100
	if percent < 0 {
		percent = 0
	} else if percent > 100 {
		percent = 100
	}

	// Calculate the width of the filled portion
	filledWidth := (width * percent) / 100

	// Create the filled and empty portions with more visually appealing characters
	filled := strings.Repeat("█", filledWidth)
	empty := strings.Repeat("▒", width-filledWidth)

	// Add a border to the progress bar
	bar := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(AccentColor).
		Padding(0, 1).
		Render(
			lipgloss.JoinHorizontal(
				lipgloss.Center,
				lipgloss.NewStyle().Foreground(PrimaryColor).Render(filled),
				lipgloss.NewStyle().Foreground(DimmedColor).Render(empty),
			),
		)

	// Add percentage text
	percentText := lipgloss.NewStyle().
		Foreground(SuccessColor).
		Bold(true).
		Render(fmt.Sprintf(" %d%% ", percent))

	return lipgloss.JoinHorizontal(lipgloss.Center, bar, percentText)
}

// Info creates an info message
func Info(content string) string {
	return lipgloss.NewStyle().
		Foreground(TextColor).
		Render(content)
}

// Success creates a success message
func Success(content string) string {
	return lipgloss.NewStyle().
		Foreground(SuccessColor).
		Bold(true).
		Render(content)
}

// Warning creates a warning message
func Warning(content string) string {
	return lipgloss.NewStyle().
		Foreground(WarningColor).
		Bold(true).
		Render(content)
}

// Error creates an error message
func Error(content string) string {
	return lipgloss.NewStyle().
		Foreground(ErrorColor).
		Bold(true).
		Render(content)
}

// Spinner creates a spinner with the given content
func Spinner(spinner, content string) string {
	spinnerStyle := lipgloss.NewStyle().
		Foreground(PrimaryColor).
		Bold(true)

	contentStyle := lipgloss.NewStyle().
		Foreground(TextColor)

	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		spinnerStyle.Render(spinner),
		" ",
		contentStyle.Render(content),
	)
}
