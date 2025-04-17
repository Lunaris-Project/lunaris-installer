package tui

import (
	"strings"

	"github.com/Lunaris-Project/lunaris-installer/pkg/tui/messages"
)

// AddInfoMessage adds an info message to the message queue
func (m *Model) AddInfoMessage(content string, source string) {
	// Add to the message queue
	if m.messageQueue != nil {
		m.messageQueue.Add(messages.NewInfoMessage(content, source))
	}

	// Also add to the legacy system messages for backward compatibility
	m.systemMessages = append(m.systemMessages, content)
}

// AddSuccessMessage adds a success message to the message queue
func (m *Model) AddSuccessMessage(content string, source string) {
	// Add to the message queue
	if m.messageQueue != nil {
		m.messageQueue.Add(messages.NewSuccessMessage(content, source))
	}

	// Also add to the legacy system messages for backward compatibility
	m.systemMessages = append(m.systemMessages, content)
}

// AddWarningMessage adds a warning message to the message queue
func (m *Model) AddWarningMessage(content string, source string) {
	// Add to the message queue
	if m.messageQueue != nil {
		m.messageQueue.Add(messages.NewWarningMessage(content, source))
	}

	// Also add to the legacy system messages for backward compatibility
	m.systemMessages = append(m.systemMessages, content)
}

// AddErrorMessage adds an error message to the message queue
func (m *Model) AddErrorMessage(content string, source string) {
	// Add to the message queue
	if m.messageQueue != nil {
		m.messageQueue.Add(messages.NewErrorMessage(content, source))
	}

	// Also add to the legacy system messages for backward compatibility
	m.systemMessages = append(m.systemMessages, content)
}

// AddDebugMessage adds a debug message to the message queue
func (m *Model) AddDebugMessage(content string, source string) {
	// Add to the message queue
	if m.messageQueue != nil {
		m.messageQueue.Add(messages.NewDebugMessage(content, source))
	}

	// Also add to the legacy system messages for backward compatibility
	m.systemMessages = append(m.systemMessages, content)
}

// AddMessage adds a message to the message queue with automatic type detection
func (m *Model) AddMessage(content string, source string) {
	// Determine message type based on content
	content = strings.TrimSpace(content)
	lowerContent := strings.ToLower(content)

	switch {
	case strings.Contains(lowerContent, "error") || 
	     strings.Contains(lowerContent, "failed") || 
	     strings.Contains(lowerContent, "conflict"):
		m.AddErrorMessage(content, source)
	case strings.Contains(lowerContent, "warning") || 
	     strings.Contains(lowerContent, "caution"):
		m.AddWarningMessage(content, source)
	case strings.Contains(lowerContent, "success") || 
	     strings.Contains(lowerContent, "complete") || 
	     strings.Contains(lowerContent, "installed"):
		m.AddSuccessMessage(content, source)
	case strings.Contains(lowerContent, "debug"):
		m.AddDebugMessage(content, source)
	default:
		m.AddInfoMessage(content, source)
	}
}

// ClearMessages clears all messages from the message queue
func (m *Model) ClearMessages() {
	// Clear the message queue
	if m.messageQueue != nil {
		m.messageQueue.Clear()
	}

	// Also clear the legacy system messages for backward compatibility
	m.systemMessages = make([]string, 0)
}
