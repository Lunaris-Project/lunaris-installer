package ui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nixev/hyprland-installer/aur"
	"github.com/nixev/hyprland-installer/config"
	"github.com/nixev/hyprland-installer/utils"
)

// InstallProgressMsg represents a progress update during installation
type InstallProgressMsg struct {
	Progress      int
	Total         int
	Current       string
	Error         string
	Complete      bool
	NeedsPassword bool
	HasConflict   bool
	ConflictMsg   string
}

// startInstallation starts the installation process
func (m *Model) startInstallation() tea.Cmd {
	return func() tea.Msg {
		// Calculate total packages to install
		total := len(config.AURPackages)
		for _, category := range m.categories {
			if selectedOptions, ok := m.selectedOptions[category.Name]; ok {
				for _, selectedOption := range selectedOptions {
					for _, option := range category.Options {
						if option.Name == selectedOption {
							total += len(option.Packages)
						}
					}
				}
			}
		}

		// Send initial progress message
		return InstallProgressMsg{
			Progress: 0,
			Total:    total,
			Current:  "Preparing installation...",
		}
	}
}

// Update handles installation progress messages
func (m *Model) handleInstallProgress(msg InstallProgressMsg) (tea.Model, tea.Cmd) {
	m.installProgress = msg.Progress
	m.installTotal = msg.Total
	m.installCurrent = msg.Current
	m.installError = msg.Error
	m.installComplete = msg.Complete

	// Check if password is needed
	if msg.NeedsPassword {
		m.awaitingPassword = true
		// Clear any previous password input
		m.passwordInput = ""
		// Default to visible password for better user experience
		m.passwordVisible = true
		return m, nil
	}

	// Check if there's a package conflict
	if msg.HasConflict {
		m.hasConflict = true
		m.conflictMessage = msg.ConflictMsg
		m.conflictChoice = true // Default to "yes"
		return m, nil
	}

	if msg.Complete {
		return m, nil
	}

	// Continue installation if no error
	if m.installError == "" && !m.installComplete {
		return m, m.continueInstallation()
	}

	return m, nil
}

// continueInstallation continues the installation process
func (m *Model) continueInstallation() tea.Cmd {
	return func() tea.Msg {
		// Check if we need to handle sudo password
		if m.installError != "" && (strings.Contains(m.installError, "password") ||
			strings.Contains(m.installError, "sudo") ||
			strings.Contains(m.installError, "[sudo]") ||
			strings.Contains(m.installError, "authentication") ||
			strings.Contains(strings.ToLower(m.installCurrent), "password for")) {
			return InstallProgressMsg{
				Progress:      m.installProgress,
				Total:         m.installTotal,
				Current:       "Sudo authentication required. Please enter your password below.",
				NeedsPassword: true,
			}
		}

		// Check if we need to handle package conflicts
		if m.hasConflict {
			// User has made a choice, continue with installation
			m.hasConflict = false

			// Create a command to handle the conflict response
			return func() tea.Msg {
				// Send the user's choice to the package manager
				response := "y\n"
				if !m.conflictChoice {
					response = "n\n"
				}

				// Write the response to stdin of the package manager
				if err := aur.SendInputToPackageManager(response); err != nil {
					return InstallProgressMsg{
						Progress: m.installProgress,
						Total:    m.installTotal,
						Current:  "Failed to respond to package conflict",
						Error:    err.Error(),
					}
				}

				// Continue with installation
				return InstallProgressMsg{
					Progress: m.installProgress,
					Total:    m.installTotal,
					Current:  "Continuing installation after conflict resolution...",
				}
			}
		}

		// Check if AUR helper is installed
		if !m.aurHelper.IsInstalled() {
			// Send progress update
			progressMsg := InstallProgressMsg{
				Progress: m.installProgress,
				Total:    m.installTotal,
				Current:  fmt.Sprintf("Installing AUR helper: %s", m.aurHelper.Name),
			}
			time.Sleep(500 * time.Millisecond) // Simulate work

			// Install AUR helper
			err := m.aurHelper.Install()
			if err != nil {
				// Check if error is related to password
				if strings.Contains(err.Error(), "password") ||
					strings.Contains(err.Error(), "sudo") ||
					strings.Contains(err.Error(), "[sudo]") ||
					strings.Contains(err.Error(), "authentication") ||
					strings.Contains(strings.ToLower(err.Error()), "password for") {
					return InstallProgressMsg{
						Progress:      m.installProgress,
						Total:         m.installTotal,
						Current:       "Sudo authentication required. Please enter your password below.",
						NeedsPassword: true,
					}
				}

				// Check if error is related to package conflicts
				if strings.Contains(err.Error(), "conflict") ||
					strings.Contains(err.Error(), "Proceed with installation") ||
					strings.Contains(err.Error(), "[Y/n]") {
					conflictMsg := err.Error()
					// Extract the conflict message
					if idx := strings.Index(conflictMsg, "conflict"); idx != -1 {
						conflictMsg = conflictMsg[idx:]
						if endIdx := strings.Index(conflictMsg, "\n"); endIdx != -1 && endIdx < len(conflictMsg) {
							conflictMsg = conflictMsg[:endIdx]
						}
					}

					return InstallProgressMsg{
						Progress:    m.installProgress,
						Total:       m.installTotal,
						Current:     "Package conflict detected",
						HasConflict: true,
						ConflictMsg: conflictMsg,
					}
				}

				// Other error
				return InstallProgressMsg{
					Progress: m.installProgress,
					Total:    m.installTotal,
					Current:  "Failed to install AUR helper",
					Error:    err.Error(),
				}
			}

			// Update progress
			m.installProgress++
			return progressMsg
		}

		// Get all packages to install
		var packages []string

		// Add AUR packages
		packages = append(packages, config.AURPackages...)

		// Add selected packages from categories
		for _, category := range m.categories {
			if selectedOptions, ok := m.selectedOptions[category.Name]; ok {
				for _, selectedOption := range selectedOptions {
					for _, option := range category.Options {
						if option.Name == selectedOption {
							packages = append(packages, option.Packages...)
						}
					}
				}
			}
		}

		// Get installed packages
		installedPackages, err := aur.GetInstalledPackages()
		if err != nil {
			return InstallProgressMsg{
				Progress: m.installProgress,
				Total:    m.installTotal,
				Current:  "Failed to get installed packages",
				Error:    err.Error(),
			}
		}

		// Filter out already installed packages
		var packagesToInstall []string
		for _, pkg := range packages {
			isInstalled := false
			for _, installedPkg := range installedPackages {
				if pkg == installedPkg {
					isInstalled = true
					break
				}
			}
			if !isInstalled {
				packagesToInstall = append(packagesToInstall, pkg)
			}
		}

		// Install packages in smaller batches to better handle errors
		batchSize := 3 // Reduced batch size for better control
		for i := 0; i < len(packagesToInstall); i += batchSize {
			end := i + batchSize
			if end > len(packagesToInstall) {
				end = len(packagesToInstall)
			}

			batch := packagesToInstall[i:end]
			if len(batch) == 0 {
				continue
			}

			// Update current package
			currentPackage := batch[0]
			if len(batch) > 1 {
				currentPackage += " and others"
			}

			// Log the current package being installed
			fmt.Printf("Installing package: %s\n", currentPackage)

			// Install batch
			err := m.aurHelper.InstallPackages(batch)
			if err != nil {
				// Check if error is related to password
				if strings.Contains(err.Error(), "password") ||
					strings.Contains(err.Error(), "sudo") ||
					strings.Contains(err.Error(), "[sudo]") ||
					strings.Contains(err.Error(), "authentication") ||
					strings.Contains(strings.ToLower(err.Error()), "password for") {
					return InstallProgressMsg{
						Progress:      m.installProgress,
						Total:         m.installTotal,
						Current:       "Sudo authentication required. Please enter your password below.",
						NeedsPassword: true,
					}
				}

				// Check if error is related to package conflicts
				if strings.Contains(err.Error(), "conflict") ||
					strings.Contains(err.Error(), "Proceed with installation") ||
					strings.Contains(err.Error(), "[Y/n]") {
					conflictMsg := err.Error()
					// Extract the conflict message
					if idx := strings.Index(conflictMsg, "conflict"); idx != -1 {
						conflictMsg = conflictMsg[idx:]
						if endIdx := strings.Index(conflictMsg, "\n"); endIdx != -1 && endIdx < len(conflictMsg) {
							conflictMsg = conflictMsg[:endIdx]
						}
					}

					return InstallProgressMsg{
						Progress:    m.installProgress,
						Total:       m.installTotal,
						Current:     "Package conflict detected",
						HasConflict: true,
						ConflictMsg: conflictMsg,
					}
				}

				// Try to install packages one by one to isolate problematic packages
				for _, pkg := range batch {
					singleErr := m.aurHelper.InstallPackages([]string{pkg})
					if singleErr != nil {
						// Skip this package and continue with others
						m.installProgress++
						continue
					}
					m.installProgress++
				}
			} else {
				// Update progress for successful batch
				m.installProgress += len(batch)
			}
		}

		// Clone the Lunaric repository if it doesn't exist
		if _, err := os.Stat("Lunaric"); os.IsNotExist(err) {
			// Send progress update
			progressMsg := InstallProgressMsg{
				Progress: m.installProgress,
				Total:    m.installTotal,
				Current:  "Cloning Lunaric repository...",
			}

			// Clone the repository
			cmd := exec.Command("git", "clone", "https://github.com/Lunaris-Project/Lunaric.git")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			if err := cmd.Run(); err != nil {
				return InstallProgressMsg{
					Progress: m.installProgress,
					Total:    m.installTotal,
					Current:  "Failed to clone Lunaric repository",
					Error:    err.Error(),
				}
			}

			// Update progress
			m.installProgress++
			return progressMsg
		}

		// Copy configuration files
		for _, dir := range config.ConfigDirs {
			// Send progress update
			progressMsg := InstallProgressMsg{
				Progress: m.installProgress,
				Total:    m.installTotal,
				Current:  fmt.Sprintf("Copying configuration files: %s", dir),
			}

			// Copy directory
			srcDir := filepath.Join("Lunaric", dir)
			dstDir := filepath.Join(os.Getenv("HOME"), dir)

			// Check if source directory exists
			if _, err := os.Stat(srcDir); os.IsNotExist(err) {
				// Create the directory in Lunaric if it doesn't exist
				if err := os.MkdirAll(srcDir, 0755); err != nil {
					return InstallProgressMsg{
						Progress: m.installProgress,
						Total:    m.installTotal,
						Current:  fmt.Sprintf("Failed to create directory: %s", srcDir),
						Error:    err.Error(),
					}
				}
			}

			// Create destination directory if it doesn't exist
			if err := os.MkdirAll(dstDir, 0755); err != nil {
				return InstallProgressMsg{
					Progress: m.installProgress,
					Total:    m.installTotal,
					Current:  fmt.Sprintf("Failed to create directory: %s", dstDir),
					Error:    err.Error(),
				}
			}

			// Copy directory contents
			err := utils.CopyDir(srcDir, dstDir)
			if err != nil {
				return InstallProgressMsg{
					Progress: m.installProgress,
					Total:    m.installTotal,
					Current:  fmt.Sprintf("Failed to copy configuration files: %s", dir),
					Error:    err.Error(),
				}
			}

			// Update progress
			m.installProgress++
			return progressMsg
		}

		// Installation complete
		return InstallProgressMsg{
			Progress: m.installTotal,
			Total:    m.installTotal,
			Current:  "Installation complete!",
			Complete: true,
		}
	}
}
