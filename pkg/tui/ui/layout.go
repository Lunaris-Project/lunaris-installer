package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// Layout represents a page layout
type Layout struct {
	width  int
	height int
}

// NewLayout creates a new layout
func NewLayout(width, height int) *Layout {
	return &Layout{
		width:  width,
		height: height,
	}
}

// SetWidth sets the width of the layout
func (l *Layout) SetWidth(width int) {
	l.width = width
}

// SetHeight sets the height of the layout
func (l *Layout) SetHeight(height int) {
	l.height = height
}

// CenteredPage creates a centered page layout
func (l *Layout) CenteredPage(content string) string {
	return lipgloss.NewStyle().
		Width(l.width).
		Height(l.height).
		Align(lipgloss.Center).
		AlignVertical(lipgloss.Center).
		Render(content)
}

// HeaderBodyFooter creates a header-body-footer layout
func (l *Layout) HeaderBodyFooter(header, body, footer string) string {
	// Calculate the height of the body
	headerHeight := lipgloss.Height(header)
	footerHeight := lipgloss.Height(footer)
	bodyHeight := l.height - headerHeight - footerHeight

	// Render the body with the calculated height
	renderedBody := lipgloss.NewStyle().
		Width(l.width).
		Height(bodyHeight).
		Render(body)

	// Join the header, body, and footer
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		renderedBody,
		footer,
	)
}

// SidebarContent creates a sidebar-content layout
func (l *Layout) SidebarContent(sidebar, content string, sidebarWidth int) string {
	// Calculate the width of the content
	contentWidth := l.width - sidebarWidth

	// Render the sidebar and content with the calculated widths
	renderedSidebar := lipgloss.NewStyle().
		Width(sidebarWidth).
		Height(l.height).
		Render(sidebar)

	renderedContent := lipgloss.NewStyle().
		Width(contentWidth).
		Height(l.height).
		Render(content)

	// Join the sidebar and content
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		renderedSidebar,
		renderedContent,
	)
}

// TwoColumn creates a two-column layout
func (l *Layout) TwoColumn(left, right string) string {
	// Calculate the width of each column
	columnWidth := l.width / 2

	// Render the left and right columns with the calculated widths
	renderedLeft := lipgloss.NewStyle().
		Width(columnWidth).
		Height(l.height).
		Render(left)

	renderedRight := lipgloss.NewStyle().
		Width(columnWidth).
		Height(l.height).
		Render(right)

	// Join the left and right columns
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		renderedLeft,
		renderedRight,
	)
}

// ThreeColumn creates a three-column layout
func (l *Layout) ThreeColumn(left, center, right string) string {
	// Calculate the width of each column
	columnWidth := l.width / 3

	// Render the left, center, and right columns with the calculated widths
	renderedLeft := lipgloss.NewStyle().
		Width(columnWidth).
		Height(l.height).
		Render(left)

	renderedCenter := lipgloss.NewStyle().
		Width(columnWidth).
		Height(l.height).
		Render(center)

	renderedRight := lipgloss.NewStyle().
		Width(columnWidth).
		Height(l.height).
		Render(right)

	// Join the left, center, and right columns
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		renderedLeft,
		renderedCenter,
		renderedRight,
	)
}

// Grid creates a grid layout
func (l *Layout) Grid(cells []string, columns int) string {
	// Calculate the width of each cell
	cellWidth := l.width / columns

	// Create rows
	rows := make([]string, 0)
	row := make([]string, 0)

	for i, cell := range cells {
		// Render the cell with the calculated width
		renderedCell := lipgloss.NewStyle().
			Width(cellWidth).
			Render(cell)

		row = append(row, renderedCell)

		// If we've reached the end of a row or the end of the cells, add the row to the rows
		if (i+1)%columns == 0 || i == len(cells)-1 {
			// Join the cells in the row
			renderedRow := lipgloss.JoinHorizontal(
				lipgloss.Top,
				row...,
			)

			rows = append(rows, renderedRow)
			row = make([]string, 0)
		}
	}

	// Join the rows
	return lipgloss.JoinVertical(
		lipgloss.Left,
		rows...,
	)
}
