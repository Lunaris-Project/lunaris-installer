package tui

import (
	"time"

	"github.com/Lunaris-Project/lunaris-installer/pkg/tui/ui"
	tea "github.com/charmbracelet/bubbletea"
)

// TaskMsg is a message for task updates
type TaskMsg struct {
	Name     string
	Progress int
	Total    int
	Status   string
	IsActive bool
	IsDone   bool
	HasError bool
}

// AddTask adds a task to the model
func (m *Model) AddTask(name string, total int) {
	// Create a new task
	task := ui.TaskProgress{
		Name:     name,
		Progress: 0,
		Total:    total,
		Status:   "Pending",
		IsActive: false,
		IsDone:   false,
		HasError: false,
	}

	// Add the task to the model
	m.tasks = append(m.tasks, task)
}

// UpdateTask updates a task in the model
func (m *Model) UpdateTask(name string, progress int, status string, isActive bool, isDone bool, hasError bool) tea.Cmd {
	return func() tea.Msg {
		return TaskMsg{
			Name:     name,
			Progress: progress,
			Total:    0, // Will be filled from existing task
			Status:   status,
			IsActive: isActive,
			IsDone:   isDone,
			HasError: hasError,
		}
	}
}

// StartTask starts a task
func (m *Model) StartTask(name string) tea.Cmd {
	return m.UpdateTask(name, 0, "In progress", true, false, false)
}

// CompleteTask completes a task
func (m *Model) CompleteTask(name string) tea.Cmd {
	return m.UpdateTask(name, 0, "Done", false, true, false)
}

// FailTask marks a task as failed
func (m *Model) FailTask(name string, errorMsg string) tea.Cmd {
	return m.UpdateTask(name, 0, errorMsg, false, false, true)
}

// UpdateTaskProgress updates the progress of a task
func (m *Model) UpdateTaskProgress(name string, progress int, status string) tea.Cmd {
	return m.UpdateTask(name, progress, status, true, false, false)
}

// handleTaskMsg handles a task message
func (m *Model) handleTaskMsg(msg TaskMsg) (tea.Model, tea.Cmd) {
	// Find the task
	for i, task := range m.tasks {
		if task.Name == msg.Name {
			// Update the task
			task.Progress = msg.Progress
			if msg.Total > 0 {
				task.Total = msg.Total
			}
			task.Status = msg.Status
			task.IsActive = msg.IsActive
			task.IsDone = msg.IsDone
			task.HasError = msg.HasError

			// Update the task in the model
			m.tasks[i] = task
			break
		}
	}

	return m, nil
}

// renderTasks renders the task list
func (m Model) renderTasks() string {
	return ui.TaskList(m.tasks, m.width)
}

// updateIndeterminateProgress updates the indeterminate progress position
func (m *Model) updateIndeterminateProgress() {
	m.indeterminatePos++
}

// tickIndeterminateProgress returns a command to tick the indeterminate progress
func (m Model) tickIndeterminateProgress() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return indeterminateProgressTickMsg{}
	})
}

// indeterminateProgressTickMsg is a message for indeterminate progress ticks
type indeterminateProgressTickMsg struct{}

// handleIndeterminateProgressTick handles indeterminate progress ticks
func (m Model) handleIndeterminateProgressTick() (tea.Model, tea.Cmd) {
	m.indeterminatePos++
	return m, m.tickIndeterminateProgress()
}
