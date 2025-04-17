package tui

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/Lunaris-Project/lunaris-installer/pkg/config"
	"github.com/Lunaris-Project/lunaris-installer/pkg/utils"
	tea "github.com/charmbracelet/bubbletea"
)

// startInstallation starts the installation process
func (m *Model) startInstallation() tea.Cmd {
	return func() tea.Msg {
		// Initialize the packages to install and calculate total steps
		m.packagesToInstall = m.getSelectedPackages()

		// Calculate total steps:
		// - Install AUR helper (1 step)
		// - Number of packages to install
		// - Ask for dotfiles installation (1 step)
		// - Backup directories (1 step if user chooses to backup)
		// - Clone repository (1 step)
		// - Create directories and copy files (1 step per directory)
		m.totalSteps = 1 + len(m.packagesToInstall) + 1 + 1 + 1 + len(config.ConfigDirs)
		m.installProgress = 0

		// Send initial progress message
		progressMsg := NewInstallProgressMsg(
			m.installProgress,
			m.totalSteps,
			"Starting installation...",
			"Preparation",
			nil,
		)

		// Request sudo password if needed
		m.awaitingPassword = true
		return progressMsg
	}
}

// continueInstallation continues the installation process
func (m *Model) continueInstallation() tea.Cmd {
	return func() tea.Msg {
		// If we have a conflict, handle it
		if m.hasConflict {
			switch m.conflictOption {
			case 0: // Skip
				m.packagesToInstall = m.packagesToInstall[1:]
				return m.installNextPackage()()
			case 1: // Replace
				// Continue with installation, the package manager will handle the replacement
				return m.installNextPackage()()
			case 2: // Cancel
				m.page = PackageCategoriesPage
				return nil
			}
		}

		// If we're in the dotfiles confirmation phase
		if m.installPhase == "dotfiles_confirmation" {
			if m.dotfilesConfirmation {
				// User wants to install dotfiles
				m.installPhase = "backup_confirmation"
				return NewBackupConfirmationMsg()
			} else {
				// User doesn't want to install dotfiles, skip to completion
				return NewCompleteMsg()
			}
		}

		// If we're in the backup confirmation phase
		if m.installPhase == "backup_confirmation" {
			if m.backupConfirmation {
				// User wants to backup, proceed with backup
				return m.backupConfigDirs()()
			} else {
				// User doesn't want to backup, proceed with dotfiles installation
				return m.installDotfiles()()
			}
		}

		// If we need to install the AUR helper first
		if !m.aurHelperInstalled {
			return m.installAURHelper()()
		}

		// If we have packages to install, install the next one
		if len(m.packagesToInstall) > 0 {
			return m.installNextPackage()()
		}

		// If we're done with packages, proceed to ask about dotfiles installation
		m.installPhase = "dotfiles_confirmation"
		return NewDotfilesConfirmationMsg()
	}
}

// installAURHelper installs the selected AUR helper
func (m *Model) installAURHelper() tea.Cmd {
	return func() tea.Msg {
		// Update progress for starting AUR helper installation
		m.installProgress++
		progressMsg := NewInstallProgressMsg(
			m.installProgress,
			m.totalSteps,
			fmt.Sprintf("Installing AUR helper: %s...", m.aurHelper.Name),
			"AUR Helper Installation",
			nil,
		)

		// Create a channel to send progress updates
		errorCh := make(chan error, 1)
		messagesCh := make(chan string, 10) // Buffer for messages
		doneCh := make(chan bool, 1)

		// Run the installation in a goroutine
		go func() {
			// Install the AUR helper
			messages, err := m.aurHelper.Install()

			// Send messages as they come in
			for _, msg := range messages {
				select {
				case messagesCh <- msg:
					// Message sent
				default:
					// Channel full, just continue
				}
			}

			if err != nil {
				errorCh <- err
				return
			}

			// Signal completion
			doneCh <- true
		}()

		// Set up a ticker to update the UI frequently
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		// Set up a timeout timer for long operations
		timeout := time.NewTimer(30 * time.Minute)
		defer timeout.Stop()

		// Wait for installation to complete or for progress updates
		for {
			select {
			case err := <-errorCh:
				progressMsg.Error = err
				return progressMsg

			case message := <-messagesCh:
				// Add message to message queue and system messages
				m.AddMessage(message, "aur-helper")

				// Update the current step with the message
				m.currentStep = message

				// Return a progress message to update the UI
				return NewInstallProgressMsg(
					m.installProgress,
					m.totalSteps,
					message,
					"AUR Helper Installation",
					nil,
				)

			case <-doneCh:
				// Mark AUR helper as installed
				m.aurHelperInstalled = true

				// Add final success message
				m.AddSuccessMessage(fmt.Sprintf("%s installed successfully", m.aurHelper.Name), "aur-helper")

				// Update the phase to Package Installation
				m.installPhase = "Package Installation"

				// Send a progress update to show we're moving to the next phase
				return NewInstallProgressMsg(
					m.installProgress,
					m.totalSteps,
					"Starting package installation...",
					"Package Installation",
					nil,
				)

			case <-ticker.C:
				// Check if installation is complete
				if m.aurHelper.IsInstalled() {
					// Mark AUR helper as installed
					m.aurHelperInstalled = true

					// Add final success message
					m.AddSuccessMessage(fmt.Sprintf("%s installed successfully", m.aurHelper.Name), "aur-helper")

					// Update the phase to Package Installation
					m.installPhase = "Package Installation"

					// Create progress message
					progressMsg := NewInstallProgressMsg(
						m.installProgress,
						m.totalSteps,
						"Starting package installation...",
						"Package Installation",
						nil,
					)

					// Show notification and return progress message
					return tea.Batch(
						m.AddSuccessNotification("AUR Helper Installed", fmt.Sprintf("%s has been installed successfully", m.aurHelper.Name)),
						func() tea.Msg { return progressMsg },
					)()
				}

				// Return a progress message to update the UI
				return NewInstallProgressMsg(
					m.installProgress,
					m.totalSteps,
					m.currentStep,
					"AUR Helper Installation",
					nil,
				)

			case <-timeout.C:
				// If no updates for 30 minutes, assume installation failed
				progressMsg.Error = fmt.Errorf("installation timed out after 30 minutes")
				return progressMsg
			}
		}
	}
}

// installNextPackage installs the next package
func (m *Model) installNextPackage() tea.Cmd {
	return func() tea.Msg {
		if len(m.packagesToInstall) == 0 {
			// If we're done with packages, proceed to ask about dotfiles installation
			m.installPhase = "dotfiles_confirmation"
			return NewDotfilesConfirmationMsg()
		}

		// Get the next package
		pkg := m.packagesToInstall[0]
		m.packagesToInstall = m.packagesToInstall[1:]

		// Update progress
		m.installProgress++
		progressMsg := NewInstallProgressMsg(
			m.installProgress,
			m.totalSteps,
			fmt.Sprintf("Installing %s...", pkg),
			"Package Installation",
			nil,
		)

		// Install the package
		messages, err := m.aurHelper.InstallPackages([]string{pkg})

		// Add messages to message queue and system messages
		if len(messages) > 0 {
			// Add each message to the message queue
			for _, msg := range messages {
				m.AddMessage(msg, "package-install")
			}

			// Update the current step with the last message
			if len(messages) > 0 {
				m.currentStep = messages[len(messages)-1]
			}
		}

		if err != nil {
			// Check if it's a conflict error
			if strings.Contains(err.Error(), "conflict") {
				// Extract the package name from the conflict message
				conflictPkg := pkg
				m.conflictPackage = conflictPkg

				// If we've chosen to replace all packages, automatically replace
				if m.replaceAllPackages {
					// Add a message about automatically replacing
					m.AddInfoMessage(fmt.Sprintf("Automatically replacing conflicting package: %s", conflictPkg), "conflict-resolution")

					// Continue with installation
					time.Sleep(500 * time.Millisecond) // Small delay for UI
					return m.installNextPackage()()
				}

				// Otherwise, show conflict resolution dialog
				m.hasConflict = true
				m.conflictOption = 0 // Default to Skip
				return NewConflictMsg(err.Error())
			}

			progressMsg.Error = err
			return progressMsg
		}

		// If there are more packages, continue installation
		if len(m.packagesToInstall) > 0 {
			time.Sleep(500 * time.Millisecond) // Small delay for UI
			return m.installNextPackage()()
		}

		// If we're done with packages, proceed to ask about dotfiles installation
		m.installPhase = "dotfiles_confirmation"
		return NewDotfilesConfirmationMsg()
	}
}

// backupConfigDirs backs up the user's .config and .local directories
func (m *Model) backupConfigDirs() tea.Cmd {
	return func() tea.Msg {
		// Update progress
		m.installProgress++
		progressMsg := NewInstallProgressMsg(
			m.installProgress,
			m.totalSteps,
			"Backing up configuration directories...",
			"Backup",
			nil,
		)

		homeDir, err := os.UserHomeDir()
		if err != nil {
			progressMsg.Error = fmt.Errorf("failed to get home directory: %w", err)
			return progressMsg
		}

		// Create the backup directory
		backupDir := filepath.Join(homeDir, "HyprLuna-User-Bak")
		backupMsg := fmt.Sprintf("Creating backup directory: %s", backupDir)
		m.AddInfoMessage(backupMsg, "backup")
		m.currentStep = backupMsg

		// Send an update to the UI
		progressMsg = NewInstallProgressMsg(
			m.installProgress,
			m.totalSteps,
			backupMsg,
			"Backup",
			nil,
		)

		// Create a channel to send progress updates with a small buffer
		updateCh := make(chan string, 5)

		// Create a goroutine to process updates and send them to the UI
		go func() {
			for msg := range updateCh {
				// Add message to message queue and system messages
				m.AddMessage(msg, "backup")
				m.currentStep = msg

				// Sleep briefly to allow UI updates to be processed
				time.Sleep(100 * time.Millisecond)
			}
		}()

		err = os.MkdirAll(backupDir, 0755)
		if err != nil {
			progressMsg.Error = fmt.Errorf("failed to create backup directory: %w", err)
			close(updateCh)
			return progressMsg
		}

		// Directories to backup
		dirsToBackup := []struct {
			source      string
			destination string
			exists      bool
		}{
			{".config", ".config", false},
			{".local", ".local", false},
			{".ags", ".ags", false},
		}

		// Check which directories exist
		for i, dir := range dirsToBackup {
			sourceDir := filepath.Join(homeDir, dir.source)
			if _, err := os.Stat(sourceDir); err == nil {
				dirsToBackup[i].exists = true
				updateCh <- fmt.Sprintf("Found directory to backup: %s", dir.source)
			} else {
				updateCh <- fmt.Sprintf("Directory does not exist, will skip: %s", dir.source)
			}
		}

		// Backup each directory if it exists
		for _, dir := range dirsToBackup {
			if !dir.exists {
				continue
			}

			sourceDir := filepath.Join(homeDir, dir.source)
			destDir := filepath.Join(backupDir, dir.destination)

			backupMsg := fmt.Sprintf("Backing up %s to %s", dir.source, dir.destination)
			updateCh <- backupMsg

			// Create parent directories if needed
			err = os.MkdirAll(filepath.Dir(destDir), 0755)
			if err != nil {
				progressMsg.Error = fmt.Errorf("failed to create backup directory for %s: %w", dir.source, err)
				close(updateCh)
				return progressMsg
			}

			// Use rsync-like approach for copying to reduce memory usage
			// This copies files one by one instead of loading entire directories into memory
			err = utils.CopyDirWithLowMemory(sourceDir, destDir)
			if err != nil {
				progressMsg.Error = fmt.Errorf("failed to backup %s directory: %w", dir.source, err)
				close(updateCh)
				return progressMsg
			}

			updateCh <- fmt.Sprintf("Successfully backed up %s to %s", dir.source, dir.destination)
		}

		updateCh <- "Backup completed successfully"
		close(updateCh)

		// Sleep briefly to allow final updates to be processed
		time.Sleep(500 * time.Millisecond)

		// Update the phase to Post-Installation
		m.installPhase = "Post-Installation"

		// Send a progress update to show we're moving to the next phase
		progressMsg = NewInstallProgressMsg(
			m.installProgress,
			m.totalSteps,
			"Starting dotfiles installation...",
			"Post-Installation",
			nil,
		)

		// Proceed with dotfiles installation
		return m.installDotfiles()()
	}
}

// installDotfiles installs the dotfiles
func (m *Model) installDotfiles() tea.Cmd {
	return func() tea.Msg {
		// Update progress
		m.installProgress++
		progressMsg := NewInstallProgressMsg(
			m.installProgress,
			m.totalSteps,
			"Installing dotfiles...",
			"Post-Installation",
			nil,
		)

		// Create a channel to send progress updates with a small buffer
		updateCh := make(chan string, 5)

		// Create a goroutine to process updates and send them to the UI
		go func() {
			for msg := range updateCh {
				// Add message to message queue and system messages
				m.AddMessage(msg, "dotfiles")
				m.currentStep = msg

				// Sleep briefly to allow UI updates to be processed
				time.Sleep(100 * time.Millisecond)
			}
		}()

		updateCh <- "Starting dotfiles installation..."

		// Get home directory
		homeDir, err := os.UserHomeDir()
		if err != nil {
			progressMsg.Error = fmt.Errorf("failed to get home directory: %w", err)
			close(updateCh)
			return progressMsg
		}

		// Clone the repository to ~/HyprLuna
		updateCh <- fmt.Sprintf("Cloning configuration repository from %s", config.ConfigRepo)

		// Create the HyprLuna directory in the user's home directory
		hyprLunaDir := filepath.Join(homeDir, "HyprLuna")

		// Remove the directory if it already exists
		if _, err := os.Stat(hyprLunaDir); err == nil {
			updateCh <- fmt.Sprintf("Removing existing directory: %s", hyprLunaDir)
			err = os.RemoveAll(hyprLunaDir)
			if err != nil {
				progressMsg.Error = fmt.Errorf("failed to remove existing HyprLuna directory: %w", err)
				close(updateCh)
				return progressMsg
			}
		}

		// Clone the repository with output capture but use a pipe to reduce memory usage
		cmd := exec.Command("git", "clone", "--depth=1", "--single-branch", config.ConfigRepo, hyprLunaDir)

		// Set up pipes for stdout and stderr
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			progressMsg.Error = fmt.Errorf("failed to create stdout pipe: %w", err)
			close(updateCh)
			return progressMsg
		}

		stderr, err := cmd.StderrPipe()
		if err != nil {
			progressMsg.Error = fmt.Errorf("failed to create stderr pipe: %w", err)
			close(updateCh)
			return progressMsg
		}

		updateCh <- "Running git clone command..."

		// Start the command
		if err := cmd.Start(); err != nil {
			progressMsg.Error = fmt.Errorf("failed to start git clone: %w", err)
			close(updateCh)
			return progressMsg
		}

		// Process output in a separate goroutine
		outputDone := make(chan struct{})
		go func() {
			defer close(outputDone)

			// Use scanners to read output line by line
			stdoutScanner := bufio.NewScanner(stdout)
			stderrScanner := bufio.NewScanner(stderr)

			// Process stdout
			go func() {
				for stdoutScanner.Scan() {
					line := stdoutScanner.Text()
					if line != "" {
						updateCh <- line
					}
				}
			}()

			// Process stderr
			for stderrScanner.Scan() {
				line := stderrScanner.Text()
				if line != "" {
					updateCh <- line
				}
			}
		}()

		// Wait for the command to complete
		err = cmd.Wait()

		// Wait for output processing to complete
		<-outputDone

		if err != nil {
			progressMsg.Error = fmt.Errorf("git clone failed: %v", err)
			close(updateCh)
			return progressMsg
		}

		// Check if the clone was successful by verifying directory contents
		files, err := os.ReadDir(hyprLunaDir)
		if err != nil || len(files) == 0 {
			progressMsg.Error = fmt.Errorf("repository cloned but appears to be empty")
			close(updateCh)
			return progressMsg
		}

		updateCh <- "Repository cloned successfully"

		// Get list of directories to copy
		updateCh <- "Checking which configuration directories exist in the repository..."

		// Check which directories exist in the repository
		existingDirs := []string{}
		for _, configDir := range config.ConfigDirs {
			sourceDir := filepath.Join(hyprLunaDir, configDir)
			if _, err := os.Stat(sourceDir); !os.IsNotExist(err) {
				existingDirs = append(existingDirs, configDir)
				updateCh <- fmt.Sprintf("Found directory in repository: %s", configDir)
			} else {
				updateCh <- fmt.Sprintf("Directory not found in repository, will skip: %s", configDir)
			}
		}

		// Copy configuration files from the cloned repository to the user's home directory
		for _, configDir := range existingDirs {
			updateCh <- fmt.Sprintf("Copying %s...", configDir)

			// Create the target directory
			targetDir := filepath.Join(homeDir, configDir)
			err := os.MkdirAll(filepath.Dir(targetDir), 0755)
			if err != nil {
				progressMsg.Error = fmt.Errorf("failed to create directory %s: %w", targetDir, err)
				close(updateCh)
				return progressMsg
			}

			// Copy the configuration files using the low memory copy function
			sourceDir := filepath.Join(hyprLunaDir, configDir)
			err = utils.CopyDirWithLowMemory(sourceDir, targetDir)
			if err != nil {
				progressMsg.Error = fmt.Errorf("failed to copy files to %s: %w", targetDir, err)
				close(updateCh)
				return progressMsg
			}

			updateCh <- fmt.Sprintf("Successfully copied %s", configDir)

			// Update progress for each directory copied
			m.installProgress++
			progressMsg = NewInstallProgressMsg(
				m.installProgress,
				m.totalSteps,
				fmt.Sprintf("Copied %s", configDir),
				"Post-Installation",
				nil,
			)
		}

		// Make scripts executable
		updateCh <- "Making scripts executable..."

		// Make hypr scripts executable
		hyprScriptsDir := filepath.Join(homeDir, ".config", "hypr", "scripts")
		if _, err := os.Stat(hyprScriptsDir); err == nil {
			chmodCmd := exec.Command("sh", "-c", fmt.Sprintf("chmod +x %s/*", hyprScriptsDir))
			chmodCmd.Run()
			updateCh <- "Made hypr scripts executable"
		}

		// Make ags scripts executable
		agsScriptsDir := filepath.Join(homeDir, ".config", "ags", "scripts", "hyprland")
		if _, err := os.Stat(agsScriptsDir); err == nil {
			chmodCmd := exec.Command("sh", "-c", fmt.Sprintf("chmod +x %s/*", agsScriptsDir))
			chmodCmd.Run()
			updateCh <- "Made ags scripts executable"
		}

		// Run wallpaper script
		wallpaperScript := filepath.Join(homeDir, ".config", "ags", "scripts", "color_generation", "wallpapers.sh")
		if _, err := os.Stat(wallpaperScript); err == nil {
			wallpaperCmd := exec.Command("sh", wallpaperScript, "-r")
			wallpaperCmd.Run()
			updateCh <- "Generated wallpaper colors"
		}

		// Add final system message
		updateCh <- "Dotfiles installation complete!"
		close(updateCh)

		// Sleep briefly to allow final updates to be processed
		time.Sleep(500 * time.Millisecond)

		// Installation complete
		return NewCompleteMsg()
	}
}

// getSelectedPackages returns a list of selected packages
func (m *Model) getSelectedPackages() []string {
	var packages []string

	// Add base packages
	packages = append(packages, config.BasePackages...)

	// Add selected packages from categories
	for categoryName, selectedOptions := range m.selectedOptions {
		for _, optionName := range selectedOptions {
			// Find the category
			for _, category := range m.categories {
				if category.Name == categoryName {
					// Find the option
					for _, option := range category.Options {
						if option.Name == optionName {
							// Add the packages
							packages = append(packages, option.Packages...)
							break
						}
					}
					break
				}
			}
		}
	}

	return packages
}

// handleInstallProgress handles installation progress messages
func (m *Model) handleInstallProgress(msg InstallProgressMsg) (tea.Model, tea.Cmd) {
	if msg.IsComplete {
		m.page = CompletePage
		return m, nil
	}

	if msg.HasConflict {
		m.hasConflict = true
		m.conflictMessage = msg.Conflict
		return m, nil
	}

	if msg.IsDotfilesConfirmation {
		m.installPhase = "dotfiles_confirmation"
		return m, nil
	}

	if msg.IsBackupConfirmation {
		m.installPhase = "backup_confirmation"
		return m, nil
	}

	if msg.Error != nil {
		m.errorMessage = msg.Error.Error()
		return m, nil
	}

	m.installProgress = msg.Progress
	m.totalSteps = msg.Total
	m.currentStep = msg.CurrentStep
	m.installPhase = msg.Phase

	// If we're awaiting password, don't continue installation yet
	if m.awaitingPassword {
		return m, nil
	}

	return m, m.continueInstallation()
}
