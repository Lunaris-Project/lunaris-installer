package messages

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Renderer renders messages
type Renderer struct {
	styles map[MessageType]lipgloss.Style
	width  int
	height int
}

// NewRenderer creates a new message renderer
func NewRenderer(width, height int) *Renderer {
	return &Renderer{
		styles: make(map[MessageType]lipgloss.Style),
		width:  width,
		height: height,
	}
}

// SetStyle sets the style for a message type
func (r *Renderer) SetStyle(msgType MessageType, style lipgloss.Style) {
	r.styles[msgType] = style
}

// SetWidth sets the width of the renderer
func (r *Renderer) SetWidth(width int) {
	r.width = width
}

// SetHeight sets the height of the renderer
func (r *Renderer) SetHeight(height int) {
	r.height = height
}

// Render renders messages
func (r *Renderer) Render(messages []Message, boxStyle lipgloss.Style) string {
	if len(messages) == 0 {
		// Render an empty box with placeholder text
		emptyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#565f89")).
			Align(lipgloss.Center).
			Width(r.width).
			Height(r.height)

		return boxStyle.Render(emptyStyle.Render("No messages to display..."))
	}

	// Format messages with styles
	formattedMessages := make([]string, 0, len(messages))
	for _, msg := range messages {
		formattedMessages = append(formattedMessages, msg.Render(r.styles))
	}

	// Join messages with newlines
	messagesText := strings.Join(formattedMessages, "\n")

	// Render the messages in a box
	return boxStyle.Render(messagesText)
}

// RenderWithTitle renders messages with a title
func (r *Renderer) RenderWithTitle(title string, messages []Message, boxStyle, titleStyle lipgloss.Style) string {
	messagesBox := r.Render(messages, boxStyle)
	renderedTitle := titleStyle.Render(title)

	return lipgloss.JoinVertical(lipgloss.Left, renderedTitle, messagesBox)
}
