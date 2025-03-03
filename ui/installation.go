package ui

import (
	"fmt"
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
		return m, nil
	}

	if msg.Complete {
		return m, nil
	}

	return m, m.continueInstallation()
}

// continueInstallation continues the installation process
func (m *Model) continueInstallation() tea.Cmd {
	return func() tea.Msg {
		// Check if we need to handle sudo password
		if m.installError != "" && (strings.Contains(m.installError, "password") ||
			strings.Contains(m.installError, "sudo") ||
			strings.Contains(m.installError, "authentication")) {
			return InstallProgressMsg{
				Progress:      m.installProgress,
				Total:         m.installTotal,
				Current:       "Sudo authentication required. Please enter your password below.",
				NeedsPassword: true,
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
					strings.Contains(err.Error(), "authentication") {
					return InstallProgressMsg{
						Progress:      m.installProgress,
						Total:         m.installTotal,
						Current:       "Sudo authentication required. Please enter your password below.",
						NeedsPassword: true,
					}
				}

				progressMsg.Error = fmt.Sprintf("Failed to install AUR helper: %v", err)
				return progressMsg
			}

			// Update progress
			progressMsg.Progress++
			return progressMsg
		}

		// Install AUR packages
		if m.installProgress < len(config.AURPackages) {
			pkg := config.AURPackages[m.installProgress]

			// Send progress update
			progressMsg := InstallProgressMsg{
				Progress: m.installProgress,
				Total:    m.installTotal,
				Current:  fmt.Sprintf("Installing package: %s", pkg),
			}
			time.Sleep(500 * time.Millisecond) // Simulate work

			// Check if package is already installed
			if aur.IsPackageInstalled(pkg) {
				progressMsg.Progress++
				return progressMsg
			}

			// Install package
			err := m.aurHelper.InstallPackages([]string{pkg})
			if err != nil {
				progressMsg.Error = fmt.Sprintf("Failed to install package %s: %v", pkg, err)
				// Continue anyway
				progressMsg.Progress++
				return progressMsg
			}

			// Update progress
			progressMsg.Progress++
			return progressMsg
		}

		// Install selected packages from categories
		progress := len(config.AURPackages)
		for _, category := range m.categories {
			if selectedOptions, ok := m.selectedOptions[category.Name]; ok {
				for _, selectedOption := range selectedOptions {
					for _, option := range category.Options {
						if option.Name == selectedOption {
							for _, pkg := range option.Packages {
								if progress == m.installProgress {
									// Send progress update
									progressMsg := InstallProgressMsg{
										Progress: progress,
										Total:    m.installTotal,
										Current:  fmt.Sprintf("Installing package: %s", pkg),
									}
									time.Sleep(500 * time.Millisecond) // Simulate work

									// Check if package is already installed
									if aur.IsPackageInstalled(pkg) {
										progressMsg.Progress++
										return progressMsg
									}

									// Install package
									err := m.aurHelper.InstallPackages([]string{pkg})
									if err != nil {
										progressMsg.Error = fmt.Sprintf("Failed to install package %s: %v", pkg, err)
										// Continue anyway
										progressMsg.Progress++
										return progressMsg
									}

									// Update progress
									progressMsg.Progress++
									return progressMsg
								}
								progress++
							}
						}
					}
				}
			}
		}

		// Copy configuration files
		progressMsg := InstallProgressMsg{
			Progress: m.installTotal - 1,
			Total:    m.installTotal,
			Current:  "Copying configuration files...",
		}
		time.Sleep(500 * time.Millisecond) // Simulate work

		// Get repository path
		repoPath, err := utils.GetRepoPath()
		if err != nil {
			progressMsg.Error = fmt.Sprintf("Failed to get repository path: %v", err)
			progressMsg.Progress = m.installTotal
			progressMsg.Complete = true
			return progressMsg
		}

		// Copy configuration files
		err = utils.CopyConfigDirs(repoPath, config.ConfigDirs)
		if err != nil {
			progressMsg.Error = fmt.Sprintf("Failed to copy configuration files: %v", err)
			progressMsg.Progress = m.installTotal
			progressMsg.Complete = true
			return progressMsg
		}

		// Installation complete
		progressMsg.Progress = m.installTotal
		progressMsg.Current = "Installation complete!"
		progressMsg.Complete = true
		return progressMsg
	}
}
