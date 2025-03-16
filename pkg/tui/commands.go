package tui

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nixev/hyprland-installer/pkg/config"
	"github.com/nixev/hyprland-installer/pkg/utils"
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
		doneCh := make(chan bool, 1)

		// Run the installation in a goroutine
		go func() {
			// Install the AUR helper
			err := m.aurHelper.Install()
			if err != nil {
				errorCh <- err
				return
			}

			// Signal completion
			doneCh <- true
		}()

		// Set up a ticker to update the spinner
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		// Set up a timeout timer
		timeout := time.NewTimer(30 * time.Second)
		defer timeout.Stop()

		// Wait for installation to complete or for progress updates
		for {
			select {
			case err := <-errorCh:
				progressMsg.Error = err
				return progressMsg

			case <-doneCh:
				// Mark AUR helper as installed
				m.aurHelperInstalled = true

				// Continue with package installation
				return m.continueInstallation()()

			case <-ticker.C:
				// Check if installation is complete
				if m.aurHelper.IsInstalled() {
					// Mark AUR helper as installed
					m.aurHelperInstalled = true

					// Continue with package installation
					return m.continueInstallation()()
				}

			case <-timeout.C:
				// If no updates for 30 seconds, assume installation is still in progress
				// but send an update to refresh the UI
				progressMsg = NewInstallProgressMsg(
					m.installProgress,
					m.totalSteps,
					fmt.Sprintf("Still installing %s... (this may take a while)", m.aurHelper.Name),
					"AUR Helper Installation",
					nil,
				)
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
		err := m.aurHelper.InstallPackages([]string{pkg})
		if err != nil {
			// Check if it's a conflict error
			if strings.Contains(err.Error(), "conflict") {
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
		backupDir := filepath.Join(homeDir, "Lunaric-User-Backup")
		err = os.MkdirAll(backupDir, 0755)
		if err != nil {
			progressMsg.Error = fmt.Errorf("failed to create backup directory: %w", err)
			return progressMsg
		}

		// Directories to backup
		dirsToBackup := []struct {
			source      string
			destination string
		}{
			{".config", ".config.bak"},
			{".local", ".local.bak"},
			{".ags", ".ags.bak"},
			{".fonts", ".fonts.bak"},
			{"Pictures", "Pictures.bak"},
		}

		// Backup each directory if it exists
		for _, dir := range dirsToBackup {
			sourceDir := filepath.Join(homeDir, dir.source)
			destDir := filepath.Join(backupDir, dir.destination)

			// Check if the source directory exists
			if _, err := os.Stat(sourceDir); err == nil {
				// Create parent directories if needed
				err = os.MkdirAll(filepath.Dir(destDir), 0755)
				if err != nil {
					progressMsg.Error = fmt.Errorf("failed to create backup directory for %s: %w", dir.source, err)
					return progressMsg
				}

				// Copy the directory
				err = utils.CopyDir(sourceDir, destDir)
				if err != nil {
					progressMsg.Error = fmt.Errorf("failed to backup %s directory: %w", dir.source, err)
					return progressMsg
				}

				fmt.Printf("Backed up %s to %s\n", dir.source, dir.destination)
			} else {
				fmt.Printf("Skipping backup of %s: directory does not exist\n", dir.source)
			}
		}

		// Proceed with dotfiles installation
		return m.installDotfiles()()
	}
}

// installDotfiles installs the dotfiles
func (m *Model) installDotfiles() tea.Cmd {
	return func() tea.Msg {
		return m.handlePostInstallation()()
	}
}

// handlePostInstallation handles post-installation tasks
func (m *Model) handlePostInstallation() tea.Cmd {
	return func() tea.Msg {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return NewInstallProgressMsg(
				m.installProgress,
				m.totalSteps,
				"Failed to get home directory",
				"Post-Installation",
				err,
			)
		}

		// Clone the repository to ~/Lunaric
		m.installProgress++
		progressMsg := NewInstallProgressMsg(
			m.installProgress,
			m.totalSteps,
			"Cloning configuration repository...",
			"Post-Installation",
			nil,
		)

		// Create the Lunaric directory in the user's home directory
		lunaricDir := filepath.Join(homeDir, "Lunaric")

		// Remove the directory if it already exists
		if _, err := os.Stat(lunaricDir); err == nil {
			err = os.RemoveAll(lunaricDir)
			if err != nil {
				progressMsg.Error = fmt.Errorf("failed to remove existing Lunaric directory: %w", err)
				return progressMsg
			}
		}

		// Clone the repository with output capture
		var stdout, stderr bytes.Buffer
		cmd := exec.Command("git", "clone", "--depth=1", config.ConfigRepo, lunaricDir)
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err = cmd.Run()

		if err != nil {
			// Provide detailed error message
			errMsg := stderr.String()
			if errMsg == "" {
				errMsg = stdout.String()
			}
			if errMsg == "" {
				errMsg = err.Error()
			}

			progressMsg.Error = fmt.Errorf("git clone failed: %s", errMsg)
			return progressMsg
		}

		// Check if the clone was successful by verifying directory contents
		files, err := os.ReadDir(lunaricDir)
		if err != nil || len(files) == 0 {
			progressMsg.Error = fmt.Errorf("repository cloned but appears to be empty")
			return progressMsg
		}

		// Copy configuration files from the cloned repository to the user's home directory
		for _, configDir := range config.ConfigDirs {
			m.installProgress++
			progressMsg := NewInstallProgressMsg(
				m.installProgress,
				m.totalSteps,
				fmt.Sprintf("Copying %s...", configDir),
				"Post-Installation",
				nil,
			)

			// Create the target directory
			targetDir := filepath.Join(homeDir, configDir)
			err := os.MkdirAll(filepath.Dir(targetDir), 0755)
			if err != nil {
				progressMsg.Error = fmt.Errorf("failed to create directory %s: %w", targetDir, err)
				return progressMsg
			}

			// Copy the configuration files
			sourceDir := filepath.Join(lunaricDir, configDir)
			if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
				// Log that we're skipping this directory
				fmt.Printf("Skipping %s: directory does not exist in the repository\n", configDir)
				continue
			}

			err = utils.CopyDir(sourceDir, targetDir)
			if err != nil {
				progressMsg.Error = fmt.Errorf("failed to copy files to %s: %w", targetDir, err)
				return progressMsg
			}
		}

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
