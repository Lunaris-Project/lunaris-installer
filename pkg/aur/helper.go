package aur

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"
)

var (
	currentCmd *exec.Cmd
	stdinPipe  io.WriteCloser
)

// GetCurrentPackageManager returns the current package manager command
func GetCurrentPackageManager() *exec.Cmd {
	return currentCmd
}

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
func (h *Helper) Install() ([]string, error) {
	// If the helper is already installed, return nil
	if h.IsInstalled() {
		return []string{"AUR helper already installed"}, nil
	}

	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "aur-helper")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to the temporary directory
	originalDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}
	if err := os.Chdir(tempDir); err != nil {
		return nil, fmt.Errorf("failed to change to temporary directory: %w", err)
	}
	defer os.Chdir(originalDir)

	// Collect system messages
	messages := []string{}
	messages = append(messages, fmt.Sprintf("Cloning %s repository...", h.Name))

	// Clone the AUR helper repository
	cloneCmd := exec.Command("git", "clone", fmt.Sprintf("https://aur.archlinux.org/%s.git", h.Name))

	// Capture output instead of sending directly to stdout/stderr
	var cloneOutput bytes.Buffer
	cloneCmd.Stdout = &cloneOutput
	cloneCmd.Stderr = &cloneOutput

	if err := cloneCmd.Run(); err != nil {
		return messages, fmt.Errorf("failed to clone repository: %w\nOutput: %s", err, cloneOutput.String())
	}

	// Add clone output to messages
	cloneOutputStr := cloneOutput.String()
	if cloneOutputStr != "" {
		messages = append(messages, cloneOutputStr)
	}
	messages = append(messages, fmt.Sprintf("Repository cloned successfully"))

	// Change to the AUR helper directory
	if err := os.Chdir(h.Name); err != nil {
		return messages, fmt.Errorf("failed to change to AUR helper directory: %w", err)
	}

	// Build and install the AUR helper
	messages = append(messages, fmt.Sprintf("Building and installing %s...", h.Name))

	// Use nice to reduce CPU priority
	var cmd *exec.Cmd
	if h.sudoPassword != "" {
		cmd = exec.Command("nice", "-n", "19", "sudo", "-S", "makepkg", "-si", "--noconfirm")
	} else {
		cmd = exec.Command("nice", "-n", "19", "makepkg", "-si", "--noconfirm")
	}

	// Capture output instead of sending directly to stdout/stderr
	var buildOutput bytes.Buffer
	cmd.Stdout = &buildOutput
	cmd.Stderr = &buildOutput

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return messages, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return messages, fmt.Errorf("failed to start makepkg: %w", err)
	}

	// Send the password if we have one
	if h.sudoPassword != "" {
		io.WriteString(stdin, h.sudoPassword+"\n")
	}
	stdin.Close()

	// Wait for the command to complete
	if err := cmd.Wait(); err != nil {
		buildOutputStr := buildOutput.String()
		messages = append(messages, buildOutputStr)
		return messages, fmt.Errorf("failed to build and install package: %w\nOutput: %s", err, buildOutputStr)
	}

	// Add build output to messages
	buildOutputStr := buildOutput.String()
	if buildOutputStr != "" {
		// Split by lines and add each line as a message
		outputLines := strings.Split(buildOutputStr, "\n")
		for _, line := range outputLines {
			if line != "" {
				messages = append(messages, line)
			}
		}
	}

	messages = append(messages, fmt.Sprintf("%s installed successfully", h.Name))
	return messages, nil
}

// InstallPackages installs packages using the AUR helper
func (h *Helper) InstallPackages(packages []string) ([]string, error) {
	if len(packages) == 0 {
		return []string{"No packages to install"}, nil
	}

	// Collect system messages
	messages := []string{}
	messages = append(messages, fmt.Sprintf("Installing packages: %v", packages))

	// Make sure any previous package manager process is cleared
	ClearPackageManager()

	// Kill any potentially hanging processes from previous attempts
	pkillCmd := exec.Command("pkill", "-9", h.Command)
	pkillCmd.Run()
	pkillPacmanCmd := exec.Command("pkill", "-9", "pacman")
	pkillPacmanCmd.Run()

	// Add a delay to ensure processes are killed
	time.Sleep(500 * time.Millisecond)

	// Build the command arguments
	args := []string{"-S", "--needed", "--noconfirm"}
	args = append(args, packages...)

	// Create a command that uses sudo directly if needed
	var cmd *exec.Cmd

	// Use nice to reduce CPU priority and use sudo with the AUR helper if we have a password
	if h.sudoPassword != "" {
		cmd = exec.Command("nice", "-n", "19", "sudo", "-S", h.Command)
		cmd.Args = append(cmd.Args, args...)
		messages = append(messages, fmt.Sprintf("Using sudo with password. Command: sudo -S %s %s", h.Command, strings.Join(args, " ")))
	} else {
		// No password provided, just use the AUR helper directly with nice
		cmd = exec.Command("nice", "-n", "19", h.Command)
		cmd.Args = append(cmd.Args, args...)
		messages = append(messages, fmt.Sprintf("No password provided. Command: %s %s", h.Command, strings.Join(args, " ")))
	}

	// Set up pipes for stdin, stdout, and stderr
	var err error
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return messages, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	// Capture output
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = io.MultiWriter(&stdoutBuf)
	cmd.Stderr = io.MultiWriter(&stderrBuf)

	// Start the command
	if err := cmd.Start(); err != nil {
		return messages, fmt.Errorf("failed to start command: %w", err)
	}

	// Set the global variables to track the current package manager process
	// We set these AFTER successfully starting the command
	currentCmd = cmd
	stdinPipe = stdin

	// If we have a sudo password and we're using sudo -S, send it
	if h.sudoPassword != "" {
		fmt.Fprintf(stdin, "%s\n", h.sudoPassword)
		messages = append(messages, "Sent sudo password to command")
	}

	// Create a channel to receive the command result
	resultCh := make(chan error, 1)

	// Create a channel to receive output updates
	outputCh := make(chan string, 100)

	// Start a goroutine to read output and send updates
	go func() {
		scanner := bufio.NewScanner(&stdoutBuf)
		for scanner.Scan() {
			outputCh <- scanner.Text()
		}
	}()

	// Start another goroutine to read stderr
	go func() {
		scanner := bufio.NewScanner(&stderrBuf)
		for scanner.Scan() {
			outputCh <- scanner.Text()
		}
	}()

	// Wait for the command to complete in a goroutine
	go func() {
		resultCh <- cmd.Wait()
		close(outputCh)
	}()

	// Collect output messages
	outputMessages := []string{}

	// Wait for the command to complete or timeout
	for {
		select {
		case err := <-resultCh:
			// Command completed
			if err != nil {
				// Check if the error is a conflict
				stdoutStr := stdoutBuf.String()
				stderrStr := stderrBuf.String()

				if strings.Contains(stdoutStr, "conflict") || strings.Contains(stderrStr, "conflict") {
					conflictMsg := extractConflictMessage(stdoutStr, stderrStr)
					messages = append(messages, outputMessages...)
					messages = append(messages, fmt.Sprintf("Conflict detected: %s", conflictMsg))
					return messages, fmt.Errorf("package conflict detected: %s", conflictMsg)
				}

				messages = append(messages, outputMessages...)
				messages = append(messages, fmt.Sprintf("Command failed: %v", err))
				return messages, fmt.Errorf("command failed: %w", err)
			}

			messages = append(messages, outputMessages...)
			messages = append(messages, "Packages installed successfully")
			return messages, nil

		case output, ok := <-outputCh:
			// Got output from the command
			if ok && output != "" {
				outputMessages = append(outputMessages, output)
				// Keep only the last 50 messages to avoid memory issues
				if len(outputMessages) > 50 {
					outputMessages = outputMessages[len(outputMessages)-50:]
				}
			}

		case <-time.After(30 * time.Minute): // Timeout after 30 minutes
			// Command timed out, kill it
			if cmd.Process != nil {
				cmd.Process.Kill()
			}
			messages = append(messages, outputMessages...)
			messages = append(messages, "Command timed out after 30 minutes")
			return messages, fmt.Errorf("command timed out after 30 minutes")
		}
	}
}

// extractConflictMessage extracts the conflict message from the output
func extractConflictMessage(stdout, stderr string) string {
	// Try to extract from stderr first
	if strings.Contains(stderr, "conflict") {
		lines := strings.Split(stderr, "\n")
		for _, line := range lines {
			if strings.Contains(line, "conflict") {
				return line
			}
		}
	}

	// Try to extract from stdout
	if strings.Contains(stdout, "conflict") {
		lines := strings.Split(stdout, "\n")
		for _, line := range lines {
			if strings.Contains(line, "conflict") {
				return line
			}
		}
	}

	// If we couldn't extract a specific message, return a generic one
	return "package conflicts detected"
}

// SendInputToPackageManager sends input to the current package manager process
func SendInputToPackageManager(input string) error {
	if stdinPipe == nil {
		return fmt.Errorf("no active package manager process")
	}

	_, err := fmt.Fprintf(stdinPipe, "%s\n", input)
	return err
}

// IsPackageManagerActive checks if a package manager process is active
func IsPackageManagerActive() bool {
	return currentCmd != nil && currentCmd.Process != nil
}

// GetInstalledPackages returns a list of installed packages
func GetInstalledPackages() ([]string, error) {
	cmd := exec.Command("pacman", "-Q")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	var packages []string
	for _, line := range strings.Split(stdout.String(), "\n") {
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) > 0 {
			packages = append(packages, parts[0])
		}
	}
	return packages, nil
}

// IsPackageInstalled checks if a package is installed
func IsPackageInstalled(pkg string) bool {
	cmd := exec.Command("pacman", "-Q", pkg)
	return cmd.Run() == nil
}

// SetSudoPassword sets the sudo password for the AUR helper
func (h *Helper) SetSudoPassword(password string) {
	h.sudoPassword = password
}

// GetSudoPassword returns the sudo password for the AUR helper
func (h *Helper) GetSudoPassword() string {
	return h.sudoPassword
}

// ClearPackageManager clears the current package manager process
func ClearPackageManager() {
	currentCmd = nil
	stdinPipe = nil
}
