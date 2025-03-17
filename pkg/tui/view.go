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

	// If help is shown, render help as a dropdown below the content
	if m.showHelp {
		helpContent := m.renderHelpDropdown()
		content = lipgloss.JoinVertical(lipgloss.Center, content, helpContent)
	}

	// Center the content horizontally
	content = m.centerHorizontally(content)

	// Center the content vertically based on terminal height
	if m.height > 0 {
		contentLines := strings.Count(content, "\n") + 1
		paddingTop := (m.height - contentLines - 1) / 2 // -1 for help hint
		if paddingTop > 0 {
			topPadding := strings.Repeat("\n", paddingTop)
			content = topPadding + content
		}
	}

	// Add help hint at the bottom
	helpHint := DimStyle.Render("Press ? for help")

	// Center the help hint horizontally
	helpHint = m.centerHorizontally(helpHint)

	return lipgloss.JoinVertical(lipgloss.Center, content, helpHint)
}

// centerHorizontally centers content horizontally in the terminal
func (m Model) centerHorizontally(content string) string {
	lines := strings.Split(content, "\n")
	centeredLines := make([]string, len(lines))

	for i, line := range lines {
		// Calculate visible width (without ANSI escape sequences)
		visibleWidth := lipgloss.Width(line)

		// Calculate padding needed
		padding := (m.width - visibleWidth) / 2
		if padding > 0 {
			centeredLines[i] = strings.Repeat(" ", padding) + line
		} else {
			centeredLines[i] = line
		}
	}

	return strings.Join(centeredLines, "\n")
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

	// Adjust box width based on terminal width
	boxWidth := min(m.width-10, 60)
	passwordField := BoxStyle.Copy().Width(boxWidth).Render(passwordDisplay)

	// Render instructions
	instructions := InfoStyle.Render("Press Enter to submit, Esc to cancel, Tab to toggle visibility")

	// Center everything horizontally
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

	// Adjust box width based on terminal width
	boxWidth := min(m.width-10, 70)
	message := BoxStyle.Copy().Width(boxWidth).Render(m.conflictMessage)

	// Render options
	options := []string{
		m.renderOption("Skip", m.conflictOption == 0),
		m.renderOption("Replace", m.conflictOption == 1),
		m.renderOption("Cancel", m.conflictOption == 2),
	}

	optionsStr := lipgloss.JoinVertical(lipgloss.Left, options...)

	// Render instructions
	instructions := InfoStyle.Render("Use Up/Down to select, Enter to confirm")

	// Center everything horizontally
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

// renderSystemMessages renders the system messages box efficiently
func (m Model) renderSystemMessages() string {
	// Calculate dynamic width based on terminal size
	boxWidth := max(min(m.width-10, 100), 40) // Min 40, max 100, or terminal width - 10

	if len(m.systemMessages) == 0 {
		// Empty box with placeholder text
		return lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(dimmedColor).
			Padding(1, 2).
			Width(boxWidth).
			Height(6).
			Align(lipgloss.Center).
			Render(DimStyle.Render("System messages will appear here..."))
	}

	// Get the last few messages (up to 5)
	startIdx := 0
	if len(m.systemMessages) > 5 {
		startIdx = len(m.systemMessages) - 5
	}

	recentMessages := m.systemMessages[startIdx:]

	// Format messages with timestamps and colors
	formattedMessages := make([]string, 0, len(recentMessages))
	for _, msg := range recentMessages {
		// Colorize based on message content
		var formattedMsg string
		switch {
		case strings.Contains(strings.ToLower(msg), "error") || strings.Contains(strings.ToLower(msg), "failed") || strings.Contains(strings.ToLower(msg), "conflict"):
			formattedMsg = ErrorStyle.Render(msg)
		case strings.Contains(strings.ToLower(msg), "success") || strings.Contains(strings.ToLower(msg), "complete"):
			formattedMsg = SuccessStyle.Render(msg)
		case strings.Contains(strings.ToLower(msg), "installing") || strings.Contains(strings.ToLower(msg), "building"):
			formattedMsg = InfoStyle.Render(msg)
		default:
			formattedMsg = BaseStyle.Render(msg)
		}
		formattedMessages = append(formattedMessages, formattedMsg)
	}

	messagesText := strings.Join(formattedMessages, "\n")

	// Create a scrollable box with system messages
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(dimmedColor).
		Padding(1, 2).
		Width(boxWidth).
		Height(6).
		Render(messagesText)
}

// renderWelcomePage renders the welcome page
func (m Model) renderWelcomePage() string {
	// Create a dynamic title with background that adapts to terminal width
	titleStyle := TitleStyle.Copy().
		Width(min(m.width, 80)).
		Align(lipgloss.Center)

	title := titleStyle.Render("Welcome to Lunaris Installer")
	subtitle := SubtitleStyle.Copy().
		Width(min(m.width, 80)).
		Align(lipgloss.Center).
		Render("This installer will help you set up Lunaris on your system")

	// Render instructions
	instructions := []string{
		"• Select an AUR helper to use for installation",
		"• Choose which packages to install",
		"• The installer will handle the rest",
	}

	// Calculate box width based on terminal width
	boxWidth := min(m.width-20, 70)

	instructionsStr := lipgloss.JoinVertical(
		lipgloss.Left,
		instructions...,
	)

	boxStyle := BoxStyle.Copy().Width(boxWidth).Align(lipgloss.Center)
	instructionsBox := boxStyle.Render(instructionsStr)

	// Render button
	button := m.renderButton("Continue", true)

	// Center everything horizontally
	return lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		subtitle,
		"",
		instructionsBox,
		"",
		button,
	)
}

// renderAURHelperPage renders the AUR helper selection page
func (m Model) renderAURHelperPage() string {
	// Create a dynamic title with background that adapts to terminal width
	titleStyle := TitleStyle.Copy().
		Width(min(m.width, 80)).
		Align(lipgloss.Center)

	title := titleStyle.Render("Select AUR Helper")
	subtitle := SubtitleStyle.Copy().
		Width(min(m.width, 80)).
		Align(lipgloss.Center).
		Render("Choose which AUR helper to use for installation")

	// Render options
	options := []string{}
	for i, helper := range m.aurHelperOptions {
		options = append(options, m.renderOption(helper, i == m.aurHelperIndex))
	}

	optionsStr := lipgloss.JoinVertical(lipgloss.Left, options...)

	// Calculate box width based on terminal width
	boxWidth := min(m.width-20, 60)
	boxStyle := BoxStyle.Copy().Width(boxWidth).Align(lipgloss.Center)
	optionsBox := boxStyle.Render(optionsStr)

	// Render instructions
	instructions := InfoStyle.Render("Use Up/Down to select, Enter to confirm, Esc to go back")

	// Center everything horizontally
	return lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		subtitle,
		"",
		optionsBox,
		"",
		instructions,
	)
}

// renderPackageCategoriesPage renders the package categories page
func (m Model) renderPackageCategoriesPage() string {
	// Create a dynamic title with background that adapts to terminal width
	titleStyle := TitleStyle.Copy().
		Width(min(m.width, 80)).
		Align(lipgloss.Center)

	title := titleStyle.Render("Select Packages")
	subtitle := SubtitleStyle.Copy().
		Width(min(m.width, 80)).
		Align(lipgloss.Center).
		Render("Choose which packages to install")

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

	// Calculate box width based on terminal width
	boxWidth := min(m.width-10, 80)
	boxStyle := BoxStyle.Copy().Width(boxWidth)
	contentBox := boxStyle.Render(content)

	// Render instructions
	var instructions string
	if m.optionIndex == -1 {
		instructions = InfoStyle.Render("Use Up/Down to navigate, Enter to select, Tab to switch to options, Right to install")
	} else {
		instructions = InfoStyle.Render("Use Up/Down to navigate, Enter to toggle, Tab to switch to categories, Esc to go back")
	}

	// Center everything horizontally
	return lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		subtitle,
		"",
		contentBox,
		"",
		instructions,
	)
}

// renderInstallationPage renders the installation page
func (m Model) renderInstallationPage() string {
	// Create a dynamic title with background that adapts to terminal width
	titleStyle := TitleStyle.Copy().
		Width(min(m.width, 80)).
		Align(lipgloss.Center)

	title := titleStyle.Render("Installing Lunaris")

	// If we're in the dotfiles confirmation phase
	if m.installPhase == "dotfiles_confirmation" {
		return m.renderDotfilesConfirmation()
	}

	// If we're in the backup confirmation phase
	if m.installPhase == "backup_confirmation" {
		return m.renderBackupConfirmation()
	}

	// Render progress
	progressPercentage := 0
	if m.totalSteps > 0 {
		progressPercentage = (m.installProgress * 100) / m.totalSteps
	}

	// Calculate progress bar width based on terminal width
	progressBarWidth := min(m.width-10, 80)

	// Create a more visually appealing progress bar
	progressBar := RenderProgressBar(progressBarWidth, progressPercentage)
	progressText := fmt.Sprintf("%d/%d (%d%%)", m.installProgress, m.totalSteps, progressPercentage)

	// Render current step with animated spinner
	var currentStep string
	if m.errorMessage != "" {
		currentStep = ErrorStyle.Render(m.errorMessage)
	} else {
		// Use a more visible spinner
		spinnerText := m.spinner.View()

		// Add some color and styling to the current step
		stepText := fmt.Sprintf("%s", m.currentStep)
		if m.installPhase == "AUR Helper Installation" {
			stepText = lipgloss.NewStyle().Foreground(primaryColor).Bold(true).Render(stepText)
		} else {
			stepText = InfoStyle.Render(stepText)
		}

		currentStep = fmt.Sprintf("%s %s", spinnerText, stepText)
	}

	// Render phase with better styling
	phaseStyle := SubtitleStyle.Copy().
		Foreground(primaryColor).
		Bold(true).
		Width(min(m.width, 80)).
		Align(lipgloss.Center)

	phase := phaseStyle.Render(m.installPhase)

	// Add a more descriptive message based on the current phase
	var phaseDescription string
	switch m.installPhase {
	case "AUR Helper Installation":
		phaseDescription = "Installing the AUR helper to enable access to the Arch User Repository"
	case "Package Installation":
		phaseDescription = "Installing selected packages from official repositories and AUR"
	case "Backup":
		phaseDescription = "Creating backups of your configuration files and directories"
	case "Post-Installation":
		phaseDescription = "Setting up configuration files and finalizing installation"
	default:
		phaseDescription = "Preparing your system"
	}

	phaseInfoStyle := InfoStyle.Copy().
		Width(min(m.width, 80)).
		Align(lipgloss.Center)

	phaseInfo := phaseInfoStyle.Render(phaseDescription)

	// Render system messages box
	systemMessagesBox := m.renderSystemMessages()

	// Center everything horizontally
	return lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		"",
		phase,
		phaseInfo,
		"",
		progressBar,
		progressText,
		"",
		currentStep,
		"",
		systemMessagesBox,
	)
}

// renderDotfilesConfirmation renders the dotfiles confirmation prompt
func (m Model) renderDotfilesConfirmation() string {
	// Create a dynamic title with background that adapts to terminal width
	titleStyle := TitleStyle.Copy().
		Width(min(m.width, 80)).
		Align(lipgloss.Center)

	title := titleStyle.Render("Dotfiles Installation")

	// Calculate box width based on terminal width
	boxWidth := min(m.width-20, 60)
	boxStyle := BoxStyle.Copy().Width(boxWidth).Align(lipgloss.Center)
	message := boxStyle.Render("Do you want to install the dotfiles?")

	// Render options
	options := []string{
		m.renderOption("Yes", m.dotfilesConfirmation),
		m.renderOption("No", !m.dotfilesConfirmation),
	}

	optionsStr := lipgloss.JoinVertical(lipgloss.Left, options...)

	// Render instructions
	instructions := InfoStyle.Render("Use Up/Down to select, Enter to confirm")

	// Center everything horizontally
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

// renderBackupConfirmation renders the backup confirmation prompt
func (m Model) renderBackupConfirmation() string {
	// Create a dynamic title with background that adapts to terminal width
	titleStyle := TitleStyle.Copy().
		Width(min(m.width, 80)).
		Align(lipgloss.Center)

	title := titleStyle.Render("Backup Configuration")

	// Calculate box width based on terminal width
	boxWidth := min(m.width-20, 70)
	boxStyle := BoxStyle.Copy().Width(boxWidth).Align(lipgloss.Center)
	message := boxStyle.Render("Do you want to backup your existing configuration directories before installing dotfiles?\n\nThe following directories will be backed up if they exist:\n• .config\n• .local\n• .ags\n• .fonts\n• Pictures\n\nBackups will be stored in ~/Lunaric-User-Backup/")

	// Render options
	options := []string{
		m.renderOption("Yes", m.backupConfirmation),
		m.renderOption("No", !m.backupConfirmation),
	}

	optionsStr := lipgloss.JoinVertical(lipgloss.Left, options...)

	// Render instructions
	instructions := InfoStyle.Render("Use Up/Down to select, Enter to confirm")

	// Center everything horizontally
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

// renderCompletePage renders the complete page
func (m Model) renderCompletePage() string {
	// Create a dynamic title with background that adapts to terminal width
	titleStyle := TitleStyle.Copy().
		Width(min(m.width, 80)).
		Align(lipgloss.Center)

	title := titleStyle.Render("Installation Complete")

	messageStyle := SuccessStyle.Copy().
		Width(min(m.width, 80)).
		Align(lipgloss.Center)

	message := messageStyle.Render("Lunaris has been successfully installed on your system!")

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

	// Calculate box width based on terminal width
	boxWidth := min(m.width-20, 70)
	boxStyle := BoxStyle.Copy().Width(boxWidth).Align(lipgloss.Center)
	instructionsBox := boxStyle.Render(instructionsStr)

	// Render button
	button := m.renderButton("Exit", true)

	// Center everything horizontally
	return lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		"",
		message,
		"",
		instructionsBox,
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

// renderHelpDropdown renders the help content as a dropdown
func (m Model) renderHelpDropdown() string {
	// Calculate box width based on terminal width
	boxWidth := min(m.width-10, 80)

	// Create a styled box for the help content
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(dimmedColor).
		Padding(1, 2).
		Width(boxWidth).
		Align(lipgloss.Left)

	// Create the help content
	var helpContent strings.Builder
	helpContent.WriteString(lipgloss.NewStyle().Bold(true).Render("Keyboard Controls:"))
	helpContent.WriteString("\n\n")

	// Add key bindings in a more readable format
	keyBindings := []struct {
		key         string
		description string
	}{
		{"↑/k", "Move up"},
		{"↓/j", "Move down"},
		{"←/h", "Move left/back"},
		{"→/l", "Move right/forward"},
		{"Enter/Space", "Select/Confirm"},
		{"Tab", "Switch focus"},
		{"Esc", "Go back"},
		{"q/Ctrl+C", "Quit"},
		{"?", "Toggle help"},
	}

	// Format key bindings in two columns
	for _, kb := range keyBindings {
		keyStyle := lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true).
			Width(15).
			Align(lipgloss.Left)

		descStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Width(boxWidth - 20)

		line := lipgloss.JoinHorizontal(
			lipgloss.Left,
			keyStyle.Render(kb.key),
			descStyle.Render(kb.description),
		)

		helpContent.WriteString(line)
		helpContent.WriteString("\n")
	}

	return boxStyle.Render(helpContent.String())
}

// Helper functions for min and max are already defined elsewhere in the codebase
// So we don't need to redefine them here
