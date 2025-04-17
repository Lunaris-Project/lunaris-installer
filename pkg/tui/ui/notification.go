package ui

import (
	"time"

	"github.com/charmbracelet/lipgloss"
)

// NotificationType represents the type of notification
type NotificationType int

const (
	InfoNotification NotificationType = iota
	SuccessNotification
	WarningNotification
	ErrorNotification
)

// Notification represents a notification
type Notification struct {
	Type      NotificationType
	Title     string
	Message   string
	CreatedAt time.Time
	Duration  time.Duration
	IsActive  bool
}

// NewNotification creates a new notification
func NewNotification(notifType NotificationType, title, message string, duration time.Duration) Notification {
	return Notification{
		Type:      notifType,
		Title:     title,
		Message:   message,
		CreatedAt: time.Now(),
		Duration:  duration,
		IsActive:  true,
	}
}

// IsExpired checks if the notification has expired
func (n *Notification) IsExpired() bool {
	if !n.IsActive {
		return true
	}

	return time.Since(n.CreatedAt) > n.Duration
}

// Dismiss dismisses the notification
func (n *Notification) Dismiss() {
	n.IsActive = false
}

// RenderNotification renders a notification
func RenderNotification(notification Notification, width int) string {
	// Calculate notification width
	notifWidth := min(width-10, 60)

	// Create styles based on notification type
	var (
		borderColor lipgloss.Color
		titleColor  lipgloss.Color
		icon        string
	)

	switch notification.Type {
	case InfoNotification:
		borderColor = PrimaryColor
		titleColor = PrimaryColor
		icon = "ℹ"
	case SuccessNotification:
		borderColor = SuccessColor
		titleColor = SuccessColor
		icon = "✓"
	case WarningNotification:
		borderColor = WarningColor
		titleColor = WarningColor
		icon = "⚠"
	case ErrorNotification:
		borderColor = ErrorColor
		titleColor = ErrorColor
		icon = "✗"
	}

	// Create title style
	titleStyle := lipgloss.NewStyle().
		Foreground(titleColor).
		Bold(true)

	// Create message style
	messageStyle := lipgloss.NewStyle().
		Foreground(TextColor)

	// Create notification box
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(1, 2).
		Width(notifWidth)

	// Combine title and message
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		titleStyle.Render(icon+" "+notification.Title),
		messageStyle.Render(notification.Message),
	)

	// Render the notification
	return boxStyle.Render(content)
}

// RenderNotifications renders a stack of notifications
func RenderNotifications(notifications []Notification, width int) string {
	if len(notifications) == 0 {
		return ""
	}

	// Render each notification
	renderedNotifications := make([]string, 0, len(notifications))
	for _, notification := range notifications {
		if notification.IsActive {
			renderedNotifications = append(renderedNotifications, RenderNotification(notification, width))
		}
	}

	// Join notifications vertically with spacing
	return lipgloss.JoinVertical(lipgloss.Left, renderedNotifications...)
}
