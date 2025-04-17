package messages

import (
	"time"

	"github.com/charmbracelet/lipgloss"
)

// MessageType represents the type of message
type MessageType int

const (
	// InfoMessage represents an informational message
	InfoMessage MessageType = iota
	// SuccessMessage represents a success message
	SuccessMessage
	// WarningMessage represents a warning message
	WarningMessage
	// ErrorMessage represents an error message
	ErrorMessage
	// DebugMessage represents a debug message
	DebugMessage
)

// Message represents a system message
type Message struct {
	Type      MessageType
	Content   string
	Timestamp time.Time
	Source    string
	ID        string
}

// NewMessage creates a new message
func NewMessage(msgType MessageType, content string, source string) Message {
	return Message{
		Type:      msgType,
		Content:   content,
		Timestamp: time.Now(),
		Source:    source,
		ID:        generateID(),
	}
}

// NewInfoMessage creates a new info message
func NewInfoMessage(content string, source string) Message {
	return NewMessage(InfoMessage, content, source)
}

// NewSuccessMessage creates a new success message
func NewSuccessMessage(content string, source string) Message {
	return NewMessage(SuccessMessage, content, source)
}

// NewWarningMessage creates a new warning message
func NewWarningMessage(content string, source string) Message {
	return NewMessage(WarningMessage, content, source)
}

// NewErrorMessage creates a new error message
func NewErrorMessage(content string, source string) Message {
	return NewMessage(ErrorMessage, content, source)
}

// NewDebugMessage creates a new debug message
func NewDebugMessage(content string, source string) Message {
	return NewMessage(DebugMessage, content, source)
}

// String returns the string representation of the message
func (m Message) String() string {
	return m.Content
}

// Render renders the message with the appropriate style
func (m Message) Render(styles map[MessageType]lipgloss.Style) string {
	if style, ok := styles[m.Type]; ok {
		return style.Render(m.Content)
	}
	return m.Content
}

// generateID generates a unique ID for the message
func generateID() string {
	return time.Now().Format("20060102150405.000000")
}
