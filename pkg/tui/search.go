package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// handleSearchInput handles search input
func (m Model) handleSearchInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		// Exit search mode
		m.searchFocused = false
		m.searchQuery = ""
		m.filteredOptions = []string{}
		return m, nil

	case tea.KeyBackspace:
		// Delete last character
		if len(m.searchQuery) > 0 {
			m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
			m.updateFilteredOptions()
		}
		return m, nil

	case tea.KeyEnter:
		// Exit search mode but keep the filter
		m.searchFocused = false
		return m, nil

	default:
		// Add character to search query
		if msg.Type == tea.KeyRunes {
			m.searchQuery += string(msg.Runes)
			m.updateFilteredOptions()
		}
		return m, nil
	}
}

// updateFilteredOptions updates the filtered options based on the search query
func (m *Model) updateFilteredOptions() {
	// If search query is empty, clear filtered options
	if m.searchQuery == "" {
		m.filteredOptions = []string{}
		return
	}

	// Get all package options from the current category
	if m.categoryIndex >= 0 && m.categoryIndex < len(m.categories) {
		category := m.categories[m.categoryIndex]
		allOptions := []string{}
		for _, option := range category.Options {
			allOptions = append(allOptions, option.Name)
		}

		// Filter options based on search query
		m.filteredOptions = []string{}
		for _, option := range allOptions {
			if containsIgnoreCase(option, m.searchQuery) {
				m.filteredOptions = append(m.filteredOptions, option)
			}
		}
	}
}

// containsIgnoreCase checks if a string contains another string, ignoring case
func containsIgnoreCase(s, substr string) bool {
	s, substr = strings.ToLower(s), strings.ToLower(substr)
	return strings.Contains(s, substr)
}
