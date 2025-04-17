package tui

import (
	"github.com/Lunaris-Project/lunaris-installer/pkg/aur"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

// Update updates the model based on the message
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Global key handlers
		switch {
		case key.Matches(msg, m.keyMap.Quit):
			return m, tea.Quit

		case key.Matches(msg, m.keyMap.Help):
			m.showHelp = !m.showHelp
			return m, nil

		case key.Matches(msg, m.keyMap.Search):
			// Toggle search focus
			m.searchFocused = !m.searchFocused
			if !m.searchFocused {
				// Clear search when exiting search mode
				m.searchQuery = ""
				m.filteredOptions = []string{}
			}
			return m, nil

		case key.Matches(msg, m.keyMap.Back):
			// Handle back navigation
			if !m.showHelp && !m.awaitingPassword && !m.hasConflict {
				return m.router.Back(m)
			}
		}

		// If help is shown, any key dismisses it
		if m.showHelp {
			m.showHelp = false
			return m, nil
		}

		// If awaiting password, handle password input
		if m.awaitingPassword {
			return m.handlePasswordInput(msg)
		}

		// If there's a conflict, handle conflict resolution
		if m.hasConflict {
			return m.handleConflictInput(msg)
		}

		// If search is focused, handle search input
		if m.searchFocused {
			return m.handleSearchInput(msg)
		}

		// Get the current route
		currentPage := m.router.CurrentPage()
		if route, ok := m.router.GetRoute(currentPage); ok {
			// Use the route's updater
			return route.Updater(msg)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width

		// Update message renderer dimensions
		m.messageRenderer.SetWidth(m.width - 10) // Subtract some padding
		m.messageRenderer.SetHeight(15)          // Fixed height for messages

		// If we're on the welcome page and just got window size, navigate to AUR helper page
		if m.router.CurrentPage() == WelcomePage && m.animating {
			// Stop animation and continue
			m.animating = false
			m.prevContent = ""
			m.nextContent = ""
		}

	case spinner.TickMsg:
		var spinnerCmd tea.Cmd
		m.spinner, spinnerCmd = m.spinner.Update(msg)
		cmds = append(cmds, spinnerCmd)

	case InstallProgressMsg:
		return m.handleInstallProgress(msg)

	case PageTransitionMsg:
		return m.handlePageTransition(msg)

	case animationTickMsg:
		return m.handleAnimationTick()

	case NotificationMsg:
		return m.handleNotification(msg)

	case dismissNotificationMsg:
		return m.handleDismissNotification(msg)

	case TaskMsg:
		return m.handleTaskMsg(msg)

	case indeterminateProgressTickMsg:
		return m.handleIndeterminateProgressTick()
	}

	// Return any batched commands
	if len(cmds) == 0 {
		return m, nil
	}

	return m, tea.Batch(cmds...)
}

// handlePasswordInput handles password input
func (m Model) handlePasswordInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		// Submit password
		m.awaitingPassword = false
		if m.aurHelper != nil {
			m.aurHelper.SetSudoPassword(m.passwordInput)
		}
		return m, m.continueInstallation()

	case tea.KeyEsc:
		// Cancel password input
		m.awaitingPassword = false
		m.passwordInput = ""
		m.page = PackageCategoriesPage
		return m, nil

	case tea.KeyBackspace:
		// Delete last character
		if len(m.passwordInput) > 0 {
			m.passwordInput = m.passwordInput[:len(m.passwordInput)-1]
		}
		return m, nil

	case tea.KeyTab:
		// Toggle password visibility
		m.passwordVisible = !m.passwordVisible
		return m, nil

	default:
		// Add character to password
		if msg.Type == tea.KeyRunes {
			m.passwordInput += string(msg.Runes)
		}
		return m, nil
	}
}

// handleConflictInput handles conflict resolution input
func (m Model) handleConflictInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyUp:
		// Navigate up through options
		if m.conflictOption > 0 {
			m.conflictOption--
		} else {
			m.conflictOption = 3 // Wrap to bottom
		}
		return m, nil

	case tea.KeyDown:
		// Navigate down through options
		if m.conflictOption < 3 {
			m.conflictOption++
		} else {
			m.conflictOption = 0 // Wrap to top
		}
		return m, nil

	case tea.KeyEnter:
		// Confirm selection
		m.hasConflict = false

		// Handle the selected option
		switch m.conflictOption {
		case 0: // Skip
			// Add to skipped packages
			if m.conflictPackage != "" {
				m.skippedPackages[m.conflictPackage] = true
			}
		case 1: // Replace
			// Do nothing, the package will be replaced
		case 2: // All
			// Set flag to replace all future conflicts
			m.replaceAllPackages = true
		case 3: // Cancel
			// Return to package selection
			m.hasConflict = false
			return m.router.Navigate(PackageCategoriesPage, m)
		}

		return m, m.continueInstallation()

	case tea.KeyEsc:
		// Cancel conflict resolution
		m.hasConflict = false
		return m.router.Navigate(PackageCategoriesPage, m)

	default:
		return m, nil
	}
}

// updateWelcomePage updates the welcome page
func (m Model) updateWelcomePage(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter, tea.KeySpace:
		// Use the router to navigate to the AUR helper page
		return m.router.Navigate(AURHelperPage, m)
	}
	return m, nil
}

// updateAURHelperPage updates the AUR helper page
func (m Model) updateAURHelperPage(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keyMap.Up):
		m.aurHelperIndex = max(0, m.aurHelperIndex-1)
	case key.Matches(msg, m.keyMap.Down):
		m.aurHelperIndex = min(len(m.aurHelperOptions)-1, m.aurHelperIndex+1)
	case key.Matches(msg, m.keyMap.Enter):
		// Set the AUR helper
		m.aurHelper = aur.NewHelper(m.aurHelperOptions[m.aurHelperIndex])

		// Initialize selected options with defaults
		for _, category := range m.categories {
			for _, option := range category.Options {
				if option.Default {
					if _, ok := m.selectedOptions[category.Name]; !ok {
						m.selectedOptions[category.Name] = []string{}
					}
					m.selectedOptions[category.Name] = append(m.selectedOptions[category.Name], option.Name)
				}
			}
		}

		// Use the router to navigate to the package categories page
		return m.router.Navigate(PackageCategoriesPage, m)
	case key.Matches(msg, m.keyMap.Back):
		// Use the router to navigate back
		return m.router.Back(m)
	}
	return m, nil
}

// updatePackageCategoriesPage updates the package categories page
func (m Model) updatePackageCategoriesPage(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keyMap.Tab):
		// Toggle focus between categories and options
		if m.optionIndex == -1 {
			m.optionIndex = 0
		} else {
			m.optionIndex = -1
		}
	case key.Matches(msg, m.keyMap.Up):
		if m.optionIndex == -1 {
			// Navigate categories
			m.categoryIndex = max(0, m.categoryIndex-1)
		} else {
			// Navigate options
			m.optionIndex = max(0, m.optionIndex-1)
		}
	case key.Matches(msg, m.keyMap.Down):
		if m.optionIndex == -1 {
			// Navigate categories
			m.categoryIndex = min(len(m.categories)-1, m.categoryIndex+1)
		} else {
			// Navigate options
			category := m.categories[m.categoryIndex]
			m.optionIndex = min(len(category.Options)-1, m.optionIndex+1)
		}
	case key.Matches(msg, m.keyMap.Enter):
		if m.optionIndex == -1 {
			// If categories are focused, switch to options
			m.optionIndex = 0
		} else {
			// Toggle option selection
			category := m.categories[m.categoryIndex]
			option := category.Options[m.optionIndex]

			// Initialize the map entry if it doesn't exist
			if _, ok := m.selectedOptions[category.Name]; !ok {
				m.selectedOptions[category.Name] = []string{}
			}

			// Check if the option is already selected
			isSelected := false
			for i, selectedOption := range m.selectedOptions[category.Name] {
				if selectedOption == option.Name {
					// Remove the option
					m.selectedOptions[category.Name] = append(
						m.selectedOptions[category.Name][:i],
						m.selectedOptions[category.Name][i+1:]...,
					)
					isSelected = true
					break
				}
			}

			// If not selected, add it
			if !isSelected {
				m.selectedOptions[category.Name] = append(m.selectedOptions[category.Name], option.Name)
			}
		}
	case key.Matches(msg, m.keyMap.Back):
		// Use the router to navigate back
		return m.router.Back(m)
	case key.Matches(msg, m.keyMap.Right):
		// Use the router to navigate to the installation page
		return m.router.Navigate(InstallationPage, m)
	}
	return m, nil
}

// updateInstallationPage updates the installation page
func (m Model) updateInstallationPage(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle dotfiles confirmation
	if m.installPhase == "dotfiles_confirmation" {
		switch msg.Type {
		case tea.KeyUp, tea.KeyDown:
			// Toggle between Yes and No
			m.dotfilesConfirmation = !m.dotfilesConfirmation
			return m, nil

		case tea.KeyEnter, tea.KeySpace:
			// Confirm selection and continue installation
			return m, m.continueInstallation()

		case tea.KeyEsc:
			// Cancel installation
			return m.router.Navigate(PackageCategoriesPage, m)
		}
	}

	// Handle backup confirmation
	if m.installPhase == "backup_confirmation" {
		switch msg.Type {
		case tea.KeyUp, tea.KeyDown:
			// Toggle between Yes and No
			m.backupConfirmation = !m.backupConfirmation
			return m, nil

		case tea.KeyEnter, tea.KeySpace:
			// Confirm selection and continue installation
			return m, m.continueInstallation()

		case tea.KeyEsc:
			// Cancel installation
			return m.router.Navigate(PackageCategoriesPage, m)
		}
	}

	// No key handlers for other installation phases
	return m, nil
}

// updateCompletePage updates the complete page
func (m Model) updateCompletePage(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keyMap.Enter):
		return m, tea.Quit
	}
	return m, nil
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
