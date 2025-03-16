package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// View renders the current view of the model
func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// If help is shown, render help
	if m.showHelp {
		return m.help.View(m.keyMap)
	}

	// If awaiting password, render password prompt
	if m.awaitingPassword {
		return m.renderPasswordPrompt()
	}

	// If there's a conflict, render conflict resolution
	if m.hasConflict {
		return m.renderConflictResolution()
	}

	// Render the current page
	var content string
	switch m.page {
	case WelcomePage:
		content = m.renderWelcomePage()
	case AURHelperPage:
		content = m.renderAURHelperPage()
	case PackageCategoriesPage:
		content = m.renderPackageCategoriesPage()
	case InstallationPage:
		content = m.renderInstallationPage()
	case CompletePage:
		content = m.renderCompletePage()
	}

	// Add help hint at the bottom
	helpHint := DimStyle.Render("Press ? for help")
	return lipgloss.JoinVertical(lipgloss.Left, content, helpHint)
}

// renderPasswordPrompt renders the password prompt
func (m Model) renderPasswordPrompt() string {
	title := TitleStyle.Render("Enter sudo password")
	subtitle := SubtitleStyle.Render("Password is required to install packages")

	// Render password field
	var passwordDisplay string
	if m.passwordVisible {
		passwordDisplay = m.passwordInput
	} else {
		passwordDisplay = strings.Repeat("*", len(m.passwordInput))
	}

	passwordField := BoxStyle.Render(passwordDisplay)

	// Render instructions
	instructions := InfoStyle.Render("Press Enter to submit, Esc to cancel, Tab to toggle visibility")

	return lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		subtitle,
		"",
		passwordField,
		"",
		instructions,
	)
}

// renderConflictResolution renders the conflict resolution dialog
func (m Model) renderConflictResolution() string {
	title := TitleStyle.Render("Package Conflict")
	message := BoxStyle.Render(m.conflictMessage)

	// Render options
	options := []string{
		m.renderOption("Skip", m.conflictOption == 0),
		m.renderOption("Replace", m.conflictOption == 1),
		m.renderOption("Cancel", m.conflictOption == 2),
	}

	optionsStr := lipgloss.JoinVertical(lipgloss.Left, options...)

	// Render instructions
	instructions := InfoStyle.Render("Use Up/Down to select, Enter to confirm")

	return lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		"",
		message,
		"",
		optionsStr,
		"",
		instructions,
	)
}

// renderSystemMessages renders the system messages box
func (m Model) renderSystemMessages() string {
	if len(m.systemMessages) == 0 {
		return ""
	}

	// Limit the number of messages to display to reduce memory usage
	maxMessages := 5
	startIdx := 0
	if len(m.systemMessages) > maxMessages {
		startIdx = len(m.systemMessages) - maxMessages
	}

	// Build the messages string more efficiently
	var b strings.Builder
	b.WriteString("\n\n")
	b.WriteString(lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#874BFD")).
		Padding(0, 1).
		Render("System Messages:"))
	b.WriteString("\n")

	// Only render the last few messages to save memory
	for i := startIdx; i < len(m.systemMessages); i++ {
		msg := m.systemMessages[i]
		style := lipgloss.NewStyle()

		// Apply color based on message content
		switch {
		case strings.Contains(strings.ToLower(msg), "error"):
			style = style.Foreground(lipgloss.Color("#FF0000"))
		case strings.Contains(strings.ToLower(msg), "warning"):
			style = style.Foreground(lipgloss.Color("#FFFF00"))
		case strings.Contains(strings.ToLower(msg), "success"):
			style = style.Foreground(lipgloss.Color("#00FF00"))
		default:
			style = style.Foreground(lipgloss.Color("#FFFFFF"))
		}

		b.WriteString(style.Render(msg))
		b.WriteString("\n")
	}

	return b.String()
}

// renderWelcomePage renders the welcome page
func (m Model) renderWelcomePage() string {
	var b strings.Builder

	b.WriteString(lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		Render("Welcome to Lunaris Installer"))

	b.WriteString("\n\n")

	b.WriteString("This installer will help you install Lunaris on your system.\n\n")

	b.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Render("Press Enter to continue..."))

	return b.String()
}

// renderAURHelperPage renders the AUR helper selection page
func (m Model) renderAURHelperPage() string {
	title := TitleStyle.Render("Select AUR Helper")
	subtitle := SubtitleStyle.Render("Choose which AUR helper to use for installation")

	// Render options
	options := []string{}
	for i, helper := range m.aurHelperOptions {
		options = append(options, m.renderOption(helper, i == m.aurHelperIndex))
	}

	optionsStr := lipgloss.JoinVertical(lipgloss.Left, options...)

	// Render instructions
	instructions := InfoStyle.Render("Use Up/Down to select, Enter to confirm, Esc to go back")

	return lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		subtitle,
		"",
		BoxStyle.Render(optionsStr),
		"",
		instructions,
	)
}

// renderPackageCategoriesPage renders the package categories page
func (m Model) renderPackageCategoriesPage() string {
	title := TitleStyle.Render("Select Packages")
	subtitle := SubtitleStyle.Render("Choose which packages to install")

	// Render categories and options
	var content string
	if len(m.categories) > 0 {
		// Render categories
		categoriesContent := []string{}
		for i, category := range m.categories {
			isSelected := i == m.categoryIndex
			isFocused := m.optionIndex == -1

			// Determine style based on selection and focus
			var categoryStyle lipgloss.Style
			if isSelected && isFocused {
				categoryStyle = SelectionStyle.Copy().Bold(true)
			} else if isSelected {
				categoryStyle = SelectionStyle
			} else {
				categoryStyle = BaseStyle
			}

			categoriesContent = append(categoriesContent, categoryStyle.Render(category.Name))

			// If this category is selected, render its options
			if isSelected {
				optionsContent := []string{}
				for j, option := range category.Options {
					isOptionSelected := j == m.optionIndex
					isFocused := m.optionIndex != -1

					// Check if this option is in the selected options
					isChecked := false
					if selectedOptions, ok := m.selectedOptions[category.Name]; ok {
						for _, selectedOption := range selectedOptions {
							if selectedOption == option.Name {
								isChecked = true
								break
							}
						}
					}

					// Determine style based on selection and focus
					var optionStyle lipgloss.Style
					if isOptionSelected && isFocused {
						optionStyle = SelectionStyle.Copy().Bold(true)
					} else {
						optionStyle = BaseStyle
					}

					// Render checkbox and option name
					checkbox := RenderCheckbox(isChecked)
					optionStr := fmt.Sprintf("%s %s", checkbox, option.Name)
					optionsContent = append(optionsContent, optionStyle.Render(optionStr))
				}

				// Indent options
				for i, option := range optionsContent {
					optionsContent[i] = "  " + option
				}

				// Add options to categories content
				categoriesContent = append(categoriesContent, optionsContent...)
			}
		}

		content = lipgloss.JoinVertical(lipgloss.Left, categoriesContent...)
	} else {
		content = InfoStyle.Render("No package categories available")
	}

	// Render instructions
	var instructions string
	if m.optionIndex == -1 {
		instructions = InfoStyle.Render("Use Up/Down to navigate, Enter to select, Tab to switch to options, Right to install")
	} else {
		instructions = InfoStyle.Render("Use Up/Down to navigate, Enter to toggle, Tab to switch to categories, Esc to go back")
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		subtitle,
		"",
		BoxStyle.Render(content),
		"",
		instructions,
	)
}

// renderInstallationPage renders the installation page
func (m Model) renderInstallationPage() string {
	var b strings.Builder

	b.WriteString(lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		Render("Installing Lunaris"))

	b.WriteString("\n\n")

	// Render progress information
	progressPercentage := 0
	if m.totalSteps > 0 {
		progressPercentage = (m.installProgress * 100) / m.totalSteps
	}

	// Create a progress bar
	progressBar := fmt.Sprintf("[%s%s] %d%%",
		strings.Repeat("=", progressPercentage/5),
		strings.Repeat(" ", 20-progressPercentage/5),
		progressPercentage)

	b.WriteString(progressBar)

	// Render system messages
	b.WriteString(m.renderSystemMessages())

	b.WriteString("\n\n")

	if m.installComplete {
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00")).
			Render("Installation complete! "))

		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Render("You can now reboot your system and select Lunaris from your display manager."))

		b.WriteString("\n\n")

		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Render("Press Ctrl+C to exit..."))
	}

	return b.String()
}

// renderCompletePage renders the complete page
func (m Model) renderCompletePage() string {
	title := TitleStyle.Render("Installation Complete")
	message := SuccessStyle.Render("Lunaris has been successfully installed on your system!")

	// Render instructions
	instructions := []string{
		"• Log out of your current session",
		"• Select Lunaris from your display manager",
		"• Your configuration files have been installed",
		"• If you chose to backup, your original files are in ~/Lunaric-User-Backup/",
		"• Enjoy your new desktop environment!",
	}

	instructionsStr := lipgloss.JoinVertical(
		lipgloss.Left,
		instructions...,
	)

	// Render button
	button := m.renderButton("Exit", true)

	return lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		"",
		message,
		"",
		BoxStyle.Render(instructionsStr),
		"",
		button,
	)
}

// renderOption renders an option with selection indicator
func (m Model) renderOption(text string, selected bool) string {
	if selected {
		return SelectionStyle.Render("> " + text)
	}
	return BaseStyle.Render("  " + text)
}

// renderButton renders a button
func (m Model) renderButton(text string, selected bool) string {
	if selected {
		return ButtonStyle.Copy().Background(primaryColor).Render(text)
	}
	return ButtonStyle.Render(text)
}
