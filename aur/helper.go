package aur

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// Global variables to track the current package manager process
var (
	currentCmd *exec.Cmd
	stdinPipe  io.WriteCloser
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

	// If we have a sudo password, use sudo -S to read from stdin
	if h.sudoPassword != "" {
		// Use a bash script to handle the sudo password
		cmd = exec.Command("bash", "-c", "echo '"+h.sudoPassword+"' | sudo -S makepkg -si --noconfirm")
	} else {
		// Use regular makepkg
		cmd = exec.Command("makepkg", "-si", "--noconfirm")
	}

	// Capture output
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = io.MultiWriter(os.Stdout, &stdoutBuf)
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)

	// Run the command
	err = cmd.Run()
	if err != nil {
		// Check if the error is due to a sudo password issue
		if strings.Contains(stderrBuf.String(), "[sudo]") ||
			strings.Contains(stderrBuf.String(), "password for") {
			return fmt.Errorf("sudo password required or incorrect: %w", err)
		}

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

	// Create a command that doesn't use sudo directly
	// Let the AUR helper handle sudo permissions itself
	cmd := exec.Command(h.Command, args...)

	// Set up pipes for stdin, stdout, and stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	// Capture output
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = io.MultiWriter(os.Stdout, &stdoutBuf)
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)

	// Start the command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	// If we have a sudo password, send it immediately
	// This will be buffered until sudo asks for it
	if h.sudoPassword != "" {
		fmt.Fprintf(stdin, "%s\n", h.sudoPassword)
	}

	// Close stdin to signal no more input
	stdin.Close()

	// Wait for the command to complete
	err = cmd.Wait()

	// Check for errors
	if err != nil {
		// Check if the error is due to a package conflict
		if strings.Contains(stdoutBuf.String(), "conflict") ||
			strings.Contains(stdoutBuf.String(), "Proceed with installation") ||
			strings.Contains(stdoutBuf.String(), "[Y/n]") {
			return fmt.Errorf("package conflict detected: %s", stdoutBuf.String())
		}

		// Check if the error is due to a package not found
		if strings.Contains(stderrBuf.String(), "target not found") {
			return fmt.Errorf("package not found: %w", err)
		}

		// Check if the error is due to a sudo password prompt
		if strings.Contains(stderrBuf.String(), "[sudo]") ||
			strings.Contains(stderrBuf.String(), "password for") {
			return fmt.Errorf("sudo password required or incorrect: %w", err)
		}

		// General error
		return fmt.Errorf("failed to install packages: %w", err)
	}

	return nil
}

// SendInputToPackageManager sends input to the current package manager process
func SendInputToPackageManager(input string) error {
	if stdinPipe == nil {
		return fmt.Errorf("no active package manager process")
	}

	_, err := stdinPipe.Write([]byte(input))
	if err != nil {
		return fmt.Errorf("failed to send input to package manager: %w", err)
	}

	return nil
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
