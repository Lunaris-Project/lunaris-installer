package tui

import (
	"fmt"
	"strings"

	"github.com/Lunaris-Project/lunaris-installer/pkg/tui/ui"
	"github.com/charmbracelet/lipgloss"
)

// View renders the current view of the model
func (m Model) View() string {
	// Create a loading screen with spinner if width is not set yet
	if m.width == 0 {
		loadingStyle := lipgloss.NewStyle().
			Foreground(ui.TextColor).
			Align(lipgloss.Center).
			AlignVertical(lipgloss.Center).
			Width(80). // Default width
			Height(24) // Default height

		spinnerText := ui.Spinner(m.spinner.View(), "Initializing...")
		return loadingStyle.Render(spinnerText)
	}

	// If awaiting password, render password prompt
	if m.awaitingPassword {
		return m.renderPasswordPrompt()
	}

	// If there's a conflict, render conflict resolution
	if m.hasConflict {
		return m.renderConflictResolution()
	}

	// If we're animating, render the animation
	if m.animating {
		// Apply animation to content
		var content string
		switch m.animation.Type {
		case ui.FadeIn, ui.SlideLeft, ui.SlideUp:
			// For these animations, we animate the new content
			content = ui.AnimateContent(m.nextContent, m.animation, m.width, m.height)
		case ui.FadeOut, ui.SlideRight, ui.SlideDown:
			// For these animations, we animate the old content
			content = ui.AnimateContent(m.prevContent, m.animation, m.width, m.height)
		default:
			// Default to just showing the next content
			content = m.nextContent
		}

		return content
	}

	// Get the current route from the router
	currentPage := m.router.CurrentPage()
	route, ok := m.router.GetRoute(currentPage)
	if !ok {
		// Create an error message if the page is not found
		errorStyle := lipgloss.NewStyle().
			Foreground(ui.ErrorColor).
			Bold(true).
			Align(lipgloss.Center).
			AlignVertical(lipgloss.Center).
			Width(m.width).
			Height(m.height)

		return errorStyle.Render("Error: Page not found - " + string(currentPage))
	}

	// Render the current page using the route's renderer
	content := route.Renderer()

	// If help is shown, render help as a dropdown below the content
	if m.showHelp {
		// Create a help box
		helpBoxStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ui.AccentColor).
			Padding(1).
			Width(m.width-4). // Subtract some padding
			Margin(1, 0, 0, 0)

		// Render help content
		helpContent := m.help.View(m.keyMap)
		helpBox := helpBoxStyle.Render(helpContent)

		// Join content and help box
		return lipgloss.JoinVertical(lipgloss.Left, content, helpBox)
	}

	// Add a footer with basic instructions
	footerStyle := lipgloss.NewStyle().
		Foreground(ui.DimmedColor).
		Align(lipgloss.Right).
		Width(m.width).
		Margin(1, 0, 0, 0)

	footer := footerStyle.Render("Press ? for help")

	// Render notifications if there are any
	notifications := m.renderNotifications()
	if notifications != "" {
		// Position notifications at the top right
		notificationsStyle := lipgloss.NewStyle().
			Align(lipgloss.Right).
			Width(m.width)

		notifications = notificationsStyle.Render(notifications)

		// Add notifications above the content
		content = lipgloss.JoinVertical(lipgloss.Left, notifications, content)
	}

	return lipgloss.JoinVertical(lipgloss.Left, content, footer)
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
	// Use our common page container style
	pageStyle := lipgloss.NewStyle().
		Width(m.width).   // Use full terminal width
		Height(m.height). // Use full terminal height
		Align(lipgloss.Center).
		AlignVertical(lipgloss.Center)

	// Create a dynamic title with background that adapts to terminal width
	titleStyle := lipgloss.NewStyle().
		Foreground(ui.PrimaryColor).
		Bold(true).
		Width(min(m.width, 80)).
		Align(lipgloss.Center)

	title := titleStyle.Render("Enter sudo password")
	subtitle := lipgloss.NewStyle().
		Foreground(ui.SecondaryColor).
		Italic(true).
		Width(min(m.width, 80)).
		Align(lipgloss.Center).
		Render("Password is required to install packages")

	// Render password field
	var passwordDisplay string
	if m.passwordVisible {
		passwordDisplay = m.passwordInput
	} else {
		passwordDisplay = strings.Repeat("*", len(m.passwordInput))
	}

	// Adjust box width based on terminal width
	boxWidth := min(m.width-10, 60)

	// Create a box for the password field
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ui.AccentColor).
		Padding(1, 2).
		Width(boxWidth).
		Align(lipgloss.Center)

	passwordField := boxStyle.Render(passwordDisplay)

	// Render instructions
	instructions := lipgloss.NewStyle().
		Foreground(ui.TextColor).
		Render("Press Enter to submit, Esc to cancel, Tab to toggle visibility")

	// Combine the content
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		subtitle,
		"",
		passwordField,
		"",
		instructions,
	)

	// Return the centered content
	return pageStyle.Render(content)
}

// renderConflictResolution renders the conflict resolution dialog
func (m Model) renderConflictResolution() string {
	// Use our common page container style
	pageStyle := lipgloss.NewStyle().
		Width(m.width).   // Use full terminal width
		Height(m.height). // Use full terminal height
		Align(lipgloss.Center).
		AlignVertical(lipgloss.Center)

	// Create a dynamic title with background that adapts to terminal width
	titleStyle := lipgloss.NewStyle().
		Foreground(ui.ErrorColor).
		Bold(true).
		Width(min(m.width, 80)).
		Align(lipgloss.Center)

	title := titleStyle.Render("Package Conflict Detected")

	// Create a subtitle with more information
	subtitleStyle := lipgloss.NewStyle().
		Foreground(ui.SecondaryColor).
		Italic(true).
		Width(min(m.width, 80)).
		Align(lipgloss.Center)

	subtitle := subtitleStyle.Render("Please select how to resolve this conflict")

	// Adjust box width based on terminal width
	boxWidth := min(m.width-10, 70)

	// Create a box for the conflict message
	messageBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ui.ErrorColor).
		Padding(1, 2).
		Width(boxWidth).
		Align(lipgloss.Left)

	// Format the conflict message with error styling
	formattedMessage := lipgloss.NewStyle().
		Foreground(ui.TextColor).
		Render(m.conflictMessage)

	message := messageBox.Render(formattedMessage)

	// Render options with descriptions
	options := []struct {
		name        string
		description string
		selected    bool
	}{
		{"Skip", "Skip this package and continue installation", m.conflictOption == 0},
		{"Replace", "Replace the existing package with the new one", m.conflictOption == 1},
		{"All", "Replace all conflicting packages automatically", m.conflictOption == 2},
		{"Cancel", "Cancel the installation process", m.conflictOption == 3},
	}

	// Format options with descriptions
	formattedOptions := []string{}
	for _, option := range options {
		// Create option with name and description
		optionStyle := lipgloss.NewStyle().
			Width(boxWidth - 4).
			Align(lipgloss.Left)

		// Format the option name
		nameStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(ui.PrimaryColor)

		if option.selected {
			nameStyle = nameStyle.Copy().
				Foreground(ui.AccentColor).
				Underline(true)
		}

		// Format the option description
		descStyle := lipgloss.NewStyle().
			Foreground(ui.DimmedColor).
			Italic(true)

		// Combine name and description
		var nameText string
		if option.selected {
			nameText = "▶ " + option.name
		} else {
			nameText = "  " + option.name
		}

		optionText := lipgloss.JoinVertical(
			lipgloss.Left,
			nameStyle.Render(nameText),
			descStyle.Render("   "+option.description),
		)

		formattedOptions = append(formattedOptions, optionStyle.Render(optionText))
	}

	optionsStr := lipgloss.JoinVertical(lipgloss.Left, formattedOptions...)

	// Create a box for the options
	optionsBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ui.AccentColor).
		Padding(1, 2).
		Width(boxWidth).
		Align(lipgloss.Left)

	renderedOptionsBox := optionsBox.Render(optionsStr)

	// Render instructions
	instructions := lipgloss.NewStyle().
		Foreground(ui.TextColor).
		Render("Use Up/Down to select, Enter to confirm")

	// Combine the content
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		subtitle,
		"",
		message,
		"",
		renderedOptionsBox,
		"",
		instructions,
	)

	// Return the centered content
	return pageStyle.Render(content)
}

// renderSystemMessages renders the system messages box efficiently
func (m Model) renderSystemMessages() string {
	// Calculate dynamic width based on terminal size
	boxWidth := max(min(m.width-10, 100), 40) // Min 40, max 100, or terminal width - 10

	// Add a title for the messages box
	title := lipgloss.NewStyle().
		Foreground(ui.PrimaryColor).
		Bold(true).
		Render("Command Output")

	// Create box style for messages
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ui.PrimaryColor).
		Padding(1, 2).
		Width(boxWidth).
		Height(15) // Increased height for better visibility

	// If we have a message queue, use it
	if m.messageQueue != nil && m.messageQueue.Size() > 0 {
		// Get the last 15 messages
		messages := m.messageQueue.GetLast(15)

		// Render messages using the message renderer
		messagesBox := m.messageRenderer.Render(messages, boxStyle)

		return lipgloss.JoinVertical(lipgloss.Left, title, messagesBox)
	}

	// Fallback to legacy system messages
	if len(m.systemMessages) == 0 {
		// Empty box with placeholder text
		emptyBox := boxStyle.Copy().
			Align(lipgloss.Center).
			Render(lipgloss.NewStyle().
				Foreground(ui.DimmedColor).
				Render("Command output will appear here..."))

		return lipgloss.JoinVertical(lipgloss.Left, title, emptyBox)
	}

	// Get the last several messages (up to 15 for better visibility)
	startIdx := 0
	if len(m.systemMessages) > 15 {
		startIdx = len(m.systemMessages) - 15
	}

	recentMessages := m.systemMessages[startIdx:]

	// Format messages with colors
	formattedMessages := make([]string, 0, len(recentMessages))
	for _, msg := range recentMessages {
		// Colorize based on message content
		var formattedMsg string
		switch {
		case strings.Contains(strings.ToLower(msg), "error") || strings.Contains(strings.ToLower(msg), "failed") || strings.Contains(strings.ToLower(msg), "conflict"):
			formattedMsg = ErrorStyle.Render(msg)
		case strings.Contains(strings.ToLower(msg), "success") || strings.Contains(strings.ToLower(msg), "complete") || strings.Contains(strings.ToLower(msg), "installed"):
			formattedMsg = SuccessStyle.Render(msg)
		case strings.Contains(strings.ToLower(msg), "installing") || strings.Contains(strings.ToLower(msg), "building") || strings.Contains(strings.ToLower(msg), "backing up"):
			formattedMsg = InfoStyle.Render(msg)
		default:
			formattedMsg = BaseStyle.Render(msg)
		}
		formattedMessages = append(formattedMessages, formattedMsg)
	}

	messagesText := strings.Join(formattedMessages, "\n")

	// Create a scrollable box with system messages
	messagesBox := boxStyle.Render(messagesText)

	return lipgloss.JoinVertical(lipgloss.Left, title, messagesBox)
}

// renderWelcomePage renders the welcome page
func (m Model) renderWelcomePage() string {
	// Use our common page container style
	pageStyle := PageContainer.Copy().
		Width(m.width).  // Use full terminal width
		Height(m.height) // Use full terminal height

	// Create a dynamic title with background that adapts to terminal width
	titleStyle := TitleStyle.Copy().
		Width(min(m.width, 80)).
		Align(lipgloss.Center).
		Bold(true)

	title := titleStyle.Render("Welcome to HyprLuna Installer")
	subtitle := SubtitleStyle.Copy().
		Width(min(m.width, 80)).
		Align(lipgloss.Center).
		Render("A modern Hyprland desktop environment")

	// Render features with consistent styling
	features := []string{
		"• Hyprland compositor with modern UI",
		"• Carefully selected applications",
		"• Thoughtful default configuration",
		"• Easy installation and setup",
	}

	// Calculate box width based on terminal width
	boxWidth := min(m.width-20, 70)

	// Style each feature
	styledFeatures := []string{}
	for _, feature := range features {
		styledFeature := lipgloss.NewStyle().
			Foreground(textColor).
			Align(lipgloss.Left).
			Render(feature)
		styledFeatures = append(styledFeatures, styledFeature)
	}

	// Join the features with spacing
	featureList := lipgloss.JoinVertical(lipgloss.Left, styledFeatures...)

	// Create a box for the features using our common content box style
	boxStyle := ContentBox.Copy().Width(boxWidth)
	featuresBox := boxStyle.Render(featureList)

	// Render button with clear instruction
	button := m.renderButton("Press Enter to continue", true)

	// Combine the content
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		subtitle,
		"",
		featuresBox,
		"",
		button,
	)

	// Return the centered content
	return pageStyle.Render(content)
}

// renderAURHelperPage renders the AUR helper selection page
func (m Model) renderAURHelperPage() string {
	// Use our common page container style
	pageStyle := PageContainer.Copy().
		Width(m.width) // Use full terminal width

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
	boxStyle := ContentBox.Copy().Width(boxWidth)
	optionsBox := boxStyle.Render(optionsStr)

	// Render instructions
	instructions := InfoStyle.Render("Use Up/Down to select, Enter to confirm, Esc to go back")

	// Combine the content
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		subtitle,
		"",
		optionsBox,
		"",
		instructions,
	)

	// Return the centered content
	return pageStyle.Render(content)
}

// renderPackageCategoriesPage renders the package categories page
func (m Model) renderPackageCategoriesPage() string {
	// Use our common page container style
	pageStyle := PageContainer.Copy().
		Width(m.width) // Use full terminal width

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

				// If we have filtered options, show only those
				if len(m.filteredOptions) > 0 && m.searchQuery != "" {
					// Show filtered options
					for j, optionName := range m.filteredOptions {
						// Find the option in the category
						for _, option := range category.Options {
							if option.Name == optionName {
								isOptionSelected := j == m.optionIndex && m.optionIndex != -1
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

								// Render checkbox and option name with highlighted search match
								checkbox := RenderCheckbox(isChecked)
								highlightedName := ui.HighlightMatch(option.Name, m.searchQuery)
								optionStr := fmt.Sprintf("%s %s", checkbox, highlightedName)
								optionsContent = append(optionsContent, optionStyle.Render(optionStr))
								break
							}
						}
					}
				} else {
					// Show all options
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
	boxStyle := ContentBox.Copy().Width(boxWidth)
	contentBox := boxStyle.Render(content)

	// Render instructions
	var instructions string
	if m.optionIndex == -1 {
		instructions = InfoStyle.Render("Use Up/Down to navigate, Enter to select, Tab to switch to options, Right to install")
	} else {
		instructions = InfoStyle.Render("Use Up/Down to navigate, Enter to toggle, Tab to switch to categories, Esc to go back")
	}

	// Render search box
	searchBoxWidth := min(m.width-20, 40)
	searchBox := ui.SearchBox(m.searchQuery, searchBoxWidth, m.searchFocused)

	// Add search instructions if search is focused
	var searchInstructions string
	if m.searchFocused {
		searchInstructions = lipgloss.NewStyle().
			Foreground(ui.DimmedColor).
			Render("Type to search, Esc to cancel, Enter to confirm")
	}

	// Combine the content
	finalContent := lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		subtitle,
		"",
		searchBox,
		searchInstructions,
		"",
		contentBox,
		"",
		instructions,
	)

	// Return the centered content
	return pageStyle.Render(finalContent)
}

// renderInstallationPage renders the installation page
func (m Model) renderInstallationPage() string {
	// Use our common page container style
	pageStyle := PageContainer.Copy().
		Width(m.width) // Use full terminal width

	// Create a dynamic title with background that adapts to terminal width
	titleStyle := TitleStyle.Copy().
		Width(min(m.width, 80)).
		Align(lipgloss.Center).
		Bold(true)

	title := titleStyle.Render("Installing HyprLuna")

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
	progressBar := m.RenderProgressBar(progressBarWidth, progressPercentage)
	progressText := fmt.Sprintf("%d/%d (%d%%)", m.installProgress, m.totalSteps, progressPercentage)

	// Render current step with animated spinner
	var currentStep string
	if m.errorMessage != "" {
		currentStep = ErrorStyle.Render(m.errorMessage)
	} else {
		// Use a more visible spinner
		spinnerText := m.spinner.View()

		// Add some color and styling to the current step
		stepText := m.currentStep
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

	// Create a box for the progress information
	progressBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ui.PrimaryColor).
		Padding(1, 2).
		Width(min(m.width-10, 80)).
		Align(lipgloss.Center)

	// Combine the progress elements
	progressContent := lipgloss.JoinVertical(
		lipgloss.Center,
		phase,
		phaseInfo,
		"",
		progressBar,
		progressText,
		"",
		currentStep,
	)

	// Add task progress if there are any tasks
	if len(m.tasks) > 0 {
		// Create a box for the tasks
		taskBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ui.AccentColor).
			Padding(1, 2).
			Width(min(m.width-10, 80)).
			Align(lipgloss.Left)

		// Render the tasks
		taskContent := m.renderTasks()

		// Add a title for the tasks
		taskTitle := lipgloss.NewStyle().
			Foreground(ui.AccentColor).
			Bold(true).
			Render("Tasks")

		// Combine the title and tasks
		taskContent = lipgloss.JoinVertical(
			lipgloss.Left,
			taskTitle,
			"",
			taskContent,
		)

		// Render the task box
		renderedTaskBox := taskBox.Render(taskContent)

		// Add the task box to the progress content
		progressContent = lipgloss.JoinVertical(
			lipgloss.Center,
			progressContent,
			"",
			renderedTaskBox,
		)
	}

	// Render the progress box
	renderedProgressBox := progressBox.Render(progressContent)

	// Render system messages box
	systemMessagesBox := m.renderSystemMessages()

	// Combine everything
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		"",
		renderedProgressBox,
		"",
		systemMessagesBox,
	)

	// Return the centered content
	return pageStyle.Render(content)
}

// renderDotfilesConfirmation renders the dotfiles confirmation prompt
func (m Model) renderDotfilesConfirmation() string {
	// Use our common page container style
	pageStyle := PageContainer.Copy().
		Width(m.width) // Use full terminal width

	// Create a dynamic title with background that adapts to terminal width
	titleStyle := TitleStyle.Copy().
		Width(min(m.width, 80)).
		Align(lipgloss.Center).
		Bold(true)

	title := titleStyle.Render("Dotfiles Installation")

	// Calculate box width based on terminal width
	boxWidth := min(m.width-20, 60)
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(primaryColor).
		Padding(1, 2).
		Width(boxWidth).
		Align(lipgloss.Center)

	// Create the message
	messageStyle := SubtitleStyle.Copy().
		Align(lipgloss.Center)

	message := messageStyle.Render("Do you want to install the dotfiles?")

	// Render options
	options := []string{
		m.renderOption("Yes", m.dotfilesConfirmation),
		m.renderOption("No", !m.dotfilesConfirmation),
	}

	optionsStr := lipgloss.JoinVertical(lipgloss.Center, options...)

	// Render instructions
	instructions := InfoStyle.Render("Use Up/Down to select, Enter to confirm")

	// Combine the content
	confirmationContent := lipgloss.JoinVertical(
		lipgloss.Center,
		message,
		"",
		optionsStr,
		"",
		instructions,
	)

	// Render the box
	renderedBox := boxStyle.Render(confirmationContent)

	// Combine everything
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		"",
		renderedBox,
	)

	// Return the centered content
	return pageStyle.Render(content)
}

// renderBackupConfirmation renders the backup confirmation prompt
func (m Model) renderBackupConfirmation() string {
	// Use our common page container style
	pageStyle := PageContainer.Copy().
		Width(m.width) // Use full terminal width

	// Create a dynamic title with background that adapts to terminal width
	titleStyle := TitleStyle.Copy().
		Width(min(m.width, 80)).
		Align(lipgloss.Center).
		Bold(true)

	title := titleStyle.Render("Backup Configuration")

	// Calculate box width based on terminal width
	boxWidth := min(m.width-20, 70)
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(primaryColor).
		Padding(1, 2).
		Width(boxWidth).
		Align(lipgloss.Center)

	// Create the message with better formatting
	messageStyle := SubtitleStyle.Copy().
		Align(lipgloss.Center)

	messageHeader := messageStyle.Render("Do you want to backup your existing configuration directories before installing dotfiles?")

	// Format the directories list
	dirsList := []string{
		"• .config",
		"• .local",
		"• .ags",
	}

	styledDirs := []string{}
	for _, dir := range dirsList {
		styledDir := lipgloss.NewStyle().
			Foreground(textColor).
			Align(lipgloss.Left).
			Render(dir)
		styledDirs = append(styledDirs, styledDir)
	}

	dirListStr := lipgloss.JoinVertical(lipgloss.Left, styledDirs...)

	// Add the backup location info
	backupLocation := InfoStyle.Render("Backups will be stored in ~/HyprLuna-User-Bak/")

	// Render options
	options := []string{
		m.renderOption("Yes", m.backupConfirmation),
		m.renderOption("No", !m.backupConfirmation),
	}

	optionsStr := lipgloss.JoinVertical(lipgloss.Center, options...)

	// Render instructions
	instructions := InfoStyle.Render("Use Up/Down to select, Enter to confirm")

	// Combine the content
	confirmationContent := lipgloss.JoinVertical(
		lipgloss.Center,
		messageHeader,
		"",
		"The following directories will be backed up if they exist:",
		dirListStr,
		"",
		backupLocation,
		"",
		optionsStr,
		"",
		instructions,
	)

	// Render the box
	renderedBox := boxStyle.Render(confirmationContent)

	// Combine everything
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		"",
		renderedBox,
	)

	// Return the centered content
	return pageStyle.Render(content)
}

// renderCompletePage renders the complete page
func (m Model) renderCompletePage() string {
	// Use our common page container style
	pageStyle := PageContainer.Copy().
		Width(m.width) // Use full terminal width

	// Create a dynamic title with background that adapts to terminal width
	titleStyle := TitleStyle.Copy().
		Width(min(m.width, 80)).
		Align(lipgloss.Center)

	title := titleStyle.Render("Installation Complete")

	messageStyle := SuccessStyle.Copy().
		Width(min(m.width, 80)).
		Align(lipgloss.Center)

	message := messageStyle.Render("HyprLuna has been successfully installed on your system!")

	// Render instructions
	instructions := []string{
		"• Log out of your current session",
		"• Select HyprLuna from your display manager",
		"• Your configuration files have been installed",
		"• If you chose to backup, your original files are in ~/HyprLuna-User-Bak/",
		"• Enjoy your new desktop environment!",
	}

	instructionsStr := lipgloss.JoinVertical(
		lipgloss.Left,
		instructions...,
	)

	// Calculate box width based on terminal width
	boxWidth := min(m.width-20, 70)
	boxStyle := ContentBox.Copy().Width(boxWidth)
	instructionsBox := boxStyle.Render(instructionsStr)

	// Render button
	button := m.renderButton("Exit", true)

	// Combine the content
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		"",
		message,
		"",
		instructionsBox,
		"",
		button,
	)

	// Return the centered content
	return pageStyle.Render(content)
}

// renderOption renders an option with selection indicator
func (m Model) renderOption(text string, selected bool) string {
	return ui.Option(text, selected)
}

// renderButton renders a button
func (m Model) renderButton(text string, selected bool) string {
	return ui.Button(text, selected)
}

// renderHelpDropdown renders the help content as a dropdown
func (m Model) renderHelpDropdown() string {
	// Calculate box width based on terminal width
	boxWidth := min(m.width-10, 80)

	// Create a styled box for the help content
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ui.AccentColor).
		Padding(1, 2).
		Width(boxWidth).
		Align(lipgloss.Left)

	// Create the help content
	var helpContent strings.Builder
	helpContent.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ui.PrimaryColor).Render("Keyboard Controls:"))
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
			Foreground(ui.SecondaryColor).
			Bold(true).
			Width(15).
			Align(lipgloss.Left)

		descStyle := lipgloss.NewStyle().
			Foreground(ui.TextColor).
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

// RenderProgressBar renders a progress bar
func (m Model) RenderProgressBar(width, percent int) string {
	return ui.ProgressBar(width, percent)
}

// RenderIndeterminateProgressBar renders an indeterminate progress bar
func (m Model) RenderIndeterminateProgressBar(width int) string {
	return ui.IndeterminateProgressBar(width, m.indeterminatePos)
}
