package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// SearchBox creates a search box
func SearchBox(query string, width int, focused bool) string {
	// Create a box style for the search box
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(AccentColor).
		Padding(0, 1).
		Width(width)

	if focused {
		boxStyle = boxStyle.BorderForeground(PrimaryColor)
	}

	// Create a prefix for the search box
	prefix := lipgloss.NewStyle().
		Foreground(SecondaryColor).
		Render("Search: ")

	// Create a cursor if focused
	cursor := ""
	if focused {
		cursor = "_"
	}

	// Render the search box
	return boxStyle.Render(prefix + query + cursor)
}

// FilterItems filters items based on a search query
func FilterItems(items []string, query string) []string {
	if query == "" {
		return items
	}

	// Convert query to lowercase for case-insensitive search
	query = strings.ToLower(query)

	// Filter items
	filtered := []string{}
	for _, item := range items {
		if strings.Contains(strings.ToLower(item), query) {
			filtered = append(filtered, item)
		}
	}

	return filtered
}

// HighlightMatch highlights the matching part of a string
func HighlightMatch(text, query string) string {
	if query == "" {
		return text
	}

	// Convert to lowercase for case-insensitive search
	lowerText := strings.ToLower(text)
	lowerQuery := strings.ToLower(query)

	// Find the index of the query in the text
	index := strings.Index(lowerText, lowerQuery)
	if index == -1 {
		return text
	}

	// Split the text into three parts: before, match, and after
	before := text[:index]
	match := text[index : index+len(query)]
	after := text[index+len(query):]

	// Highlight the match
	highlightedMatch := lipgloss.NewStyle().
		Foreground(PrimaryColor).
		Bold(true).
		Render(match)

	// Combine the parts
	return before + highlightedMatch + after
}
