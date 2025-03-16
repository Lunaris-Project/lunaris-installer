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

// renderWelcomePage renders the welcome page
func (m Model) renderWelcomePage() string {
	title := TitleStyle.Render("Welcome to Lunaris Installer")
	subtitle := SubtitleStyle.Render("This installer will help you set up Lunaris on your system")

	// Render instructions
	instructions := []string{
		"• Select an AUR helper to use for installation",
		"• Choose which packages to install",
		"• The installer will handle the rest",
	}

	instructionsStr := lipgloss.JoinVertical(
		lipgloss.Left,
		instructions...,
	)

	// Render button
	button := m.renderButton("Continue", true)

	return lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		subtitle,
		"",
		BoxStyle.Render(instructionsStr),
		"",
		button,
	)
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
	title := TitleStyle.Render("Installing Lunaris")

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

	// Create a more visually appealing progress bar
	progressBar := RenderProgressBar(m.width-20, progressPercentage)
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
	phase := SubtitleStyle.Copy().
		Foreground(primaryColor).
		Bold(true).
		Render(m.installPhase)

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

	phaseInfo := InfoStyle.Render(phaseDescription)

	// Render system messages in a box
	var systemMessagesBox string
	if len(m.systemMessages) > 0 {
		// Get the last few messages (up to 5)
		startIdx := 0
		if len(m.systemMessages) > 5 {
			startIdx = len(m.systemMessages) - 5
		}

		recentMessages := m.systemMessages[startIdx:]

		// Format messages with timestamps and colors
		formattedMessages := []string{}
		for _, msg := range recentMessages {
			// Colorize based on message content
			var formattedMsg string
			switch {
			case strings.Contains(msg, "error") || strings.Contains(msg, "failed") || strings.Contains(msg, "conflict"):
				formattedMsg = ErrorStyle.Render(msg)
			case strings.Contains(msg, "success") || strings.Contains(msg, "complete"):
				formattedMsg = SuccessStyle.Render(msg)
			case strings.Contains(msg, "installing") || strings.Contains(msg, "building"):
				formattedMsg = InfoStyle.Render(msg)
			default:
				formattedMsg = BaseStyle.Render(msg)
			}
			formattedMessages = append(formattedMessages, formattedMsg)
		}

		messagesText := strings.Join(formattedMessages, "\n")

		// Create a scrollable box with system messages
		systemMessagesBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(dimmedColor).
			Padding(1, 2).
			Width(m.width - 20).
			Height(6).
			Render(messagesText)
	} else {
		// Empty box with placeholder text
		systemMessagesBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(dimmedColor).
			Padding(1, 2).
			Width(m.width - 20).
			Height(6).
			Render(DimStyle.Render("System messages will appear here..."))
	}

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
	title := TitleStyle.Render("Dotfiles Installation")
	message := BoxStyle.Render("Do you want to install the dotfiles?")

	// Render options
	options := []string{
		m.renderOption("Yes", m.dotfilesConfirmation),
		m.renderOption("No", !m.dotfilesConfirmation),
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

// renderBackupConfirmation renders the backup confirmation prompt
func (m Model) renderBackupConfirmation() string {
	title := TitleStyle.Render("Backup Configuration")
	message := BoxStyle.Render("Do you want to backup your existing configuration directories before installing dotfiles?\n\nThe following directories will be backed up if they exist:\n• .config\n• .local\n• .ags\n• .fonts\n• Pictures\n\nBackups will be stored in ~/Lunaric-User-Backup/")

	// Render options
	options := []string{
		m.renderOption("Yes", m.backupConfirmation),
		m.renderOption("No", !m.backupConfirmation),
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
