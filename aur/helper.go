package aur

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Helper represents an AUR helper
type Helper struct {
	Name         string
	Command      string
	sudoPassword string
}

// NewHelper creates a new AUR helper
func NewHelper(name string) *Helper {
	return &Helper{
		Name:    name,
		Command: name,
	}
}

// IsInstalled checks if the AUR helper is installed
func (h *Helper) IsInstalled() bool {
	_, err := exec.LookPath(h.Command)
	return err == nil
}

// Install installs the AUR helper
func (h *Helper) Install() error {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "aur-helper")
	if err != nil {
		return fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to the temporary directory
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}
	defer os.Chdir(currentDir)

	if err := os.Chdir(tempDir); err != nil {
		return fmt.Errorf("failed to change to temporary directory: %w", err)
	}

	// Clone the repository
	gitCmd := exec.Command("git", "clone", fmt.Sprintf("https://aur.archlinux.org/%s.git", h.Name))
	gitCmd.Stdout = os.Stdout
	gitCmd.Stderr = os.Stderr
	if err := gitCmd.Run(); err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	// Change to the repository directory
	if err := os.Chdir(h.Name); err != nil {
		return fmt.Errorf("failed to change to repository directory: %w", err)
	}

	// Build and install the package
	var cmd *exec.Cmd
	if h.sudoPassword != "" {
		// Use sudo with password
		sudoCmd := fmt.Sprintf("echo '%s' | sudo -S makepkg -si --noconfirm", h.sudoPassword)
		cmd = exec.Command("bash", "-c", sudoCmd)
	} else {
		// Use regular makepkg
		cmd = exec.Command("makepkg", "-si", "--noconfirm")
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to build and install package: %w", err)
	}

	return nil
}

// InstallPackages installs packages using the AUR helper
func (h *Helper) InstallPackages(packages []string) error {
	if len(packages) == 0 {
		return nil
	}

	// Build the command arguments
	args := []string{"-S", "--needed", "--noconfirm"}
	args = append(args, packages...)

	// Run the command
	var cmd *exec.Cmd
	if h.sudoPassword != "" {
		// Use sudo with password - improved method to handle sudo
		// First verify sudo access with the password
		verifySudoCmd := fmt.Sprintf("echo '%s' | sudo -S -v", h.sudoPassword)
		verifyCmd := exec.Command("bash", "-c", verifySudoCmd)
		verifyCmd.Stdout = os.Stdout
		verifyCmd.Stderr = os.Stderr

		if err := verifyCmd.Run(); err != nil {
			return fmt.Errorf("sudo authentication failed: %w", err)
		}

		// Now run the actual command
		sudoCmd := fmt.Sprintf("echo '%s' | sudo -S %s %s",
			h.sudoPassword,
			h.Command,
			strings.Join(args, " "))
		cmd = exec.Command("bash", "-c", sudoCmd)
	} else {
		// Use regular command
		cmd = exec.Command(h.Command, args...)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// GetInstalledPackages returns a list of installed packages
func GetInstalledPackages() ([]string, error) {
	cmd := exec.Command("pacman", "-Qq")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get installed packages: %w", err)
	}

	packages := strings.Split(strings.TrimSpace(string(output)), "\n")
	return packages, nil
}

// IsPackageInstalled checks if a package is installed
func IsPackageInstalled(pkg string) bool {
	cmd := exec.Command("pacman", "-Qq", pkg)
	err := cmd.Run()
	return err == nil
}

// SetSudoPassword sets the sudo password for the helper
func (h *Helper) SetSudoPassword(password string) {
	h.sudoPassword = password
}
