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
		// - Number of packages to install
		// - Clone repository (1 step)
		// - Create directories and copy files (1 step per directory)
		m.totalSteps = len(m.packagesToInstall) + 1 + len(config.ConfigDirs)
		m.installProgress = 0

		// Send initial progress message
		progressMsg := NewInstallProgressMsg(
			m.installProgress,
			m.totalSteps,
			"Starting installation...",
			"Preparation",
			nil,
		)

		// Check if AUR helper is installed
		if !m.aurHelper.IsInstalled() {
			// Install AUR helper
			err := m.aurHelper.Install()
			if err != nil {
				progressMsg.Error = err
				return progressMsg
			}
		}

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

		// If we have packages to install, install the next one
		if len(m.packagesToInstall) > 0 {
			return m.installNextPackage()()
		}

		// If we're done with packages, proceed to post-installation
		return m.handlePostInstallation()()
	}
}

// installNextPackage installs the next package
func (m *Model) installNextPackage() tea.Cmd {
	return func() tea.Msg {
		if len(m.packagesToInstall) == 0 {
			return m.handlePostInstallation()()
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

		// Otherwise, proceed to post-installation
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
