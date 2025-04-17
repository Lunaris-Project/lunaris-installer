package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// SimpleProgressBar creates a simple progress bar
func SimpleProgressBar(width, percent int) string {
	// Calculate the width of the filled portion
	filledWidth := int(float64(width) * float64(percent) / 100.0)

	// Ensure filledWidth is within bounds
	if filledWidth > width {
		filledWidth = width
	}
	if filledWidth < 0 {
		filledWidth = 0
	}

	// Calculate the width of the empty portion
	emptyWidth := width - filledWidth

	// Create the filled and empty portions
	filled := strings.Repeat("█", filledWidth)
	empty := strings.Repeat("░", emptyWidth)

	// Style the filled portion
	filledStyle := lipgloss.NewStyle().
		Foreground(PrimaryColor)

	// Style the empty portion
	emptyStyle := lipgloss.NewStyle().
		Foreground(DimmedColor)

	// Combine the filled and empty portions
	return filledStyle.Render(filled) + emptyStyle.Render(empty)
}

// ProgressIndicator creates a progress indicator with label and percentage
func ProgressIndicator(width, percent int, label string) string {
	// Create the progress bar
	progressBar := SimpleProgressBar(width-10, percent) // Subtract some space for the percentage

	// Create the percentage text
	percentText := fmt.Sprintf(" %3d%% ", percent)

	// Style the percentage text
	percentStyle := lipgloss.NewStyle().
		Foreground(TextColor).
		Bold(true)

	// Combine the progress bar and percentage
	progressWithPercent := lipgloss.JoinHorizontal(
		lipgloss.Center,
		progressBar,
		percentStyle.Render(percentText),
	)

	// If there's a label, add it above the progress bar
	if label != "" {
		labelStyle := lipgloss.NewStyle().
			Foreground(TextColor).
			Width(width).
			Align(lipgloss.Left)

		return lipgloss.JoinVertical(
			lipgloss.Left,
			labelStyle.Render(label),
			progressWithPercent,
		)
	}

	return progressWithPercent
}

// IndeterminateProgressBar creates an indeterminate progress bar
func IndeterminateProgressBar(width int, position int) string {
	// Calculate the position of the indicator
	pos := position % (width * 2)
	if pos > width {
		pos = width*2 - pos
	}

	// Create the bar
	bar := strings.Repeat("░", width)

	// Insert the indicator
	indicator := "█"
	if pos < width {
		bar = bar[:pos] + indicator + bar[pos+1:]
	}

	// Style the bar
	barStyle := lipgloss.NewStyle().
		Foreground(DimmedColor)

	// Style the indicator
	indicatorStyle := lipgloss.NewStyle().
		Foreground(PrimaryColor)

	// Replace the indicator with a styled version
	styledBar := barStyle.Render(bar)
	if pos < width {
		// Calculate the position in the styled string
		styledPos := pos
		for i := 0; i < pos; i++ {
			// Add the length of ANSI escape sequences for each character
			styledPos += len(barStyle.Render("░")) - 1
		}
		styledBar = styledBar[:styledPos] + indicatorStyle.Render(indicator) + styledBar[styledPos+len(indicatorStyle.Render(indicator)):]
	}

	return styledBar
}

// TaskProgress represents a task with progress
type TaskProgress struct {
	Name     string
	Progress int
	Total    int
	Status   string
	IsActive bool
	IsDone   bool
	HasError bool
}

// TaskList creates a list of tasks with progress
func TaskList(tasks []TaskProgress, width int) string {
	if len(tasks) == 0 {
		return ""
	}

	// Calculate the width of the task list
	taskWidth := width - 4 // Subtract some padding

	// Create the task list
	taskList := make([]string, 0, len(tasks))
	for _, task := range tasks {
		// Calculate the percentage
		percent := 0
		if task.Total > 0 {
			percent = (task.Progress * 100) / task.Total
		}

		// Create the task name
		nameStyle := lipgloss.NewStyle().
			Width(taskWidth / 3).
			Align(lipgloss.Left)

		var name string
		if task.IsDone {
			name = nameStyle.Copy().
				Foreground(SuccessColor).
				Render("✓ " + task.Name)
		} else if task.HasError {
			name = nameStyle.Copy().
				Foreground(ErrorColor).
				Render("✗ " + task.Name)
		} else if task.IsActive {
			name = nameStyle.Copy().
				Foreground(PrimaryColor).
				Bold(true).
				Render("▶ " + task.Name)
		} else {
			name = nameStyle.Render("  " + task.Name)
		}

		// Create the progress bar
		progressWidth := taskWidth / 3
		var progress string
		if task.IsDone {
			progress = SimpleProgressBar(progressWidth, 100)
		} else if task.HasError {
			progress = lipgloss.NewStyle().
				Foreground(ErrorColor).
				Render(strings.Repeat("!", progressWidth))
		} else if task.IsActive {
			progress = SimpleProgressBar(progressWidth, percent)
		} else {
			progress = lipgloss.NewStyle().
				Foreground(DimmedColor).
				Render(strings.Repeat("·", progressWidth))
		}

		// Create the status
		statusStyle := lipgloss.NewStyle().
			Width(taskWidth / 3).
			Align(lipgloss.Right)

		var status string
		if task.IsDone {
			status = statusStyle.Copy().
				Foreground(SuccessColor).
				Render("Done")
		} else if task.HasError {
			status = statusStyle.Copy().
				Foreground(ErrorColor).
				Render(task.Status)
		} else if task.IsActive {
			status = statusStyle.Copy().
				Foreground(PrimaryColor).
				Render(task.Status)
		} else {
			status = statusStyle.Copy().
				Foreground(DimmedColor).
				Render("Pending")
		}

		// Combine the task name, progress bar, and status
		taskItem := lipgloss.JoinHorizontal(
			lipgloss.Left,
			name,
			progress,
			status,
		)

		taskList = append(taskList, taskItem)
	}

	// Join the task list
	return lipgloss.JoinVertical(lipgloss.Left, taskList...)
}
