package tui

import (
	"time"

	"github.com/Lunaris-Project/lunaris-installer/pkg/tui/ui"
	tea "github.com/charmbracelet/bubbletea"
)

// NotificationMsg is a message for notifications
type NotificationMsg struct {
	Type    ui.NotificationType
	Title   string
	Message string
}

// AddNotification adds a notification to the model
func (m *Model) AddNotification(notifType ui.NotificationType, title, message string) tea.Cmd {
	return func() tea.Msg {
		return NotificationMsg{
			Type:    notifType,
			Title:   title,
			Message: message,
		}
	}
}

// AddInfoNotification adds an info notification
func (m *Model) AddInfoNotification(title, message string) tea.Cmd {
	return m.AddNotification(ui.InfoNotification, title, message)
}

// AddSuccessNotification adds a success notification
func (m *Model) AddSuccessNotification(title, message string) tea.Cmd {
	return m.AddNotification(ui.SuccessNotification, title, message)
}

// AddWarningNotification adds a warning notification
func (m *Model) AddWarningNotification(title, message string) tea.Cmd {
	return m.AddNotification(ui.WarningNotification, title, message)
}

// AddErrorNotification adds an error notification
func (m *Model) AddErrorNotification(title, message string) tea.Cmd {
	return m.AddNotification(ui.ErrorNotification, title, message)
}

// handleNotification handles a notification message
func (m *Model) handleNotification(msg NotificationMsg) (tea.Model, tea.Cmd) {
	// Create a new notification
	notification := ui.NewNotification(
		msg.Type,
		msg.Title,
		msg.Message,
		5*time.Second, // Default duration
	)

	// Add the notification to the model
	m.notifications = append(m.notifications, notification)

	// Create a command to dismiss the notification after its duration
	return m, tea.Tick(notification.Duration, func(t time.Time) tea.Msg {
		return dismissNotificationMsg{index: len(m.notifications) - 1}
	})
}

// dismissNotificationMsg is a message to dismiss a notification
type dismissNotificationMsg struct {
	index int
}

// handleDismissNotification handles dismissing a notification
func (m *Model) handleDismissNotification(msg dismissNotificationMsg) (tea.Model, tea.Cmd) {
	// Check if the index is valid
	if msg.index >= 0 && msg.index < len(m.notifications) {
		// Dismiss the notification
		notification := m.notifications[msg.index]
		notification.Dismiss()
		m.notifications[msg.index] = notification
	}

	return m, nil
}

// renderNotifications renders all active notifications
func (m Model) renderNotifications() string {
	// Filter active notifications
	activeNotifications := make([]ui.Notification, 0)
	for i := 0; i < len(m.notifications); i++ {
		if m.notifications[i].IsActive && !m.notifications[i].IsExpired() {
			activeNotifications = append(activeNotifications, m.notifications[i])
		}
	}

	// Render notifications
	return ui.RenderNotifications(activeNotifications, m.width)
}
