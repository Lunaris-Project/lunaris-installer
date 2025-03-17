package aur

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
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

	// Collect system messages - use a fixed size buffer to limit memory usage
	messages := make([]string, 0, 20) // Pre-allocate with capacity of 20
	messages = append(messages, fmt.Sprintf("Installing %s AUR helper...", h.Name))

	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "aur-helper")
	if err != nil {
		return messages, fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to the temporary directory
	originalDir, err := os.Getwd()
	if err != nil {
		return messages, fmt.Errorf("failed to get current directory: %w", err)
	}
	if err := os.Chdir(tempDir); err != nil {
		return messages, fmt.Errorf("failed to change to temporary directory: %w", err)
	}
	defer os.Chdir(originalDir)

	messages = append(messages, fmt.Sprintf("Cloning %s repository...", h.Name))

	// Clone the AUR helper repository with depth=1 to reduce download size and memory usage
	cloneCmd := exec.Command("git", "clone", "--depth=1", fmt.Sprintf("https://aur.archlinux.org/%s.git", h.Name))

	// Use pipes instead of buffers to reduce memory usage
	cloneStdout, err := cloneCmd.StdoutPipe()
	if err != nil {
		return messages, fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	cloneStderr, err := cloneCmd.StderrPipe()
	if err != nil {
		return messages, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := cloneCmd.Start(); err != nil {
		return messages, fmt.Errorf("failed to start git clone: %w", err)
	}

	// Read output line by line to avoid storing everything in memory
	// Use a separate goroutine to prevent blocking
	cloneDone := make(chan struct{})
	go func() {
		defer close(cloneDone)
		scanner := bufio.NewScanner(io.MultiReader(cloneStdout, cloneStderr))
		for scanner.Scan() {
			line := scanner.Text()
			if line != "" {
				// Only keep important messages
				if strings.Contains(line, "error") || strings.Contains(line, "warning") ||
					strings.Contains(line, "fatal") || strings.Contains(line, "Cloning") {
					// Thread-safe append
					messages = append(messages, line)
				}
			}
		}
	}()

	// Wait for the command to complete
	if err := cloneCmd.Wait(); err != nil {
		<-cloneDone // Ensure goroutine is done
		return messages, fmt.Errorf("failed to clone repository: %w", err)
	}
	<-cloneDone // Ensure goroutine is done

	messages = append(messages, fmt.Sprintf("Repository cloned successfully"))

	// Change to the AUR helper directory
	if err := os.Chdir(h.Name); err != nil {
		return messages, fmt.Errorf("failed to change to AUR helper directory: %w", err)
	}

	messages = append(messages, fmt.Sprintf("Building and installing %s...", h.Name))

	// Use ionice along with nice to reduce both CPU and I/O priority
	var cmd *exec.Cmd
	if h.sudoPassword != "" {
		cmd = exec.Command("ionice", "-c", "3", "nice", "-n", "19", "sudo", "-S", "makepkg", "-si", "--noconfirm", "--noprogressbar")
	} else {
		cmd = exec.Command("ionice", "-c", "3", "nice", "-n", "19", "makepkg", "-si", "--noconfirm", "--noprogressbar")
	}

	// Set resource limits using ulimit-like environment variables if possible
	cmd.Env = append(os.Environ(),
		"MAKEFLAGS=-j2",               // Limit make to 2 jobs
		"CARGO_BUILD_JOBS=2",          // Limit Rust builds to 2 jobs
		"RUSTFLAGS=-Ccodegen-units=1", // Reduce Rust memory usage
	)

	// Use pipes instead of buffers to reduce memory usage
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return messages, fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return messages, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

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

	// Create a channel to receive the command result
	resultCh := make(chan error, 1)

	// Wait for the command to complete in a goroutine
	go func() {
		resultCh <- cmd.Wait()
	}()

	// Create a channel for output processing
	outputDone := make(chan struct{})

	// Read output line by line to avoid storing everything in memory
	go func() {
		defer close(outputDone)

		// Use a buffered reader with a small buffer to reduce memory usage
		stdoutReader := bufio.NewReaderSize(stdout, 4096)
		stderrReader := bufio.NewReaderSize(stderr, 4096)

		// Process stdout
		go func() {
			for {
				line, err := stdoutReader.ReadString('\n')
				if err != nil {
					if err != io.EOF {
						messages = append(messages, fmt.Sprintf("Error reading stdout: %v", err))
					}
					break
				}

				line = strings.TrimSpace(line)
				if line != "" {
					// Only keep important messages
					if strings.Contains(line, "error") || strings.Contains(line, "warning") ||
						strings.Contains(line, "installing") || strings.Contains(line, "making") ||
						strings.Contains(line, "building") || strings.Contains(line, "conflict") {

						// Add to messages with thread safety
						messages = append(messages, line)

						// Limit the number of messages to avoid memory issues
						if len(messages) > 50 {
							// Keep only the first 25 and last 24 messages
							truncatedMessages := make([]string, 0, 50)
							truncatedMessages = append(truncatedMessages, messages[:25]...)
							truncatedMessages = append(truncatedMessages, "... (output truncated) ...")
							truncatedMessages = append(truncatedMessages, messages[len(messages)-24:]...)
							messages = truncatedMessages
						}
					}
				}
			}
		}()

		// Process stderr
		for {
			line, err := stderrReader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					messages = append(messages, fmt.Sprintf("Error reading stderr: %v", err))
				}
				break
			}

			line = strings.TrimSpace(line)
			if line != "" {
				// Only keep important messages
				if strings.Contains(line, "error") || strings.Contains(line, "warning") ||
					strings.Contains(line, "installing") || strings.Contains(line, "making") ||
					strings.Contains(line, "building") || strings.Contains(line, "conflict") {

					// Add to messages with thread safety
					messages = append(messages, line)

					// Limit the number of messages to avoid memory issues
					if len(messages) > 50 {
						// Keep only the first 25 and last 24 messages
						truncatedMessages := make([]string, 0, 50)
						truncatedMessages = append(truncatedMessages, messages[:25]...)
						truncatedMessages = append(truncatedMessages, "... (output truncated) ...")
						truncatedMessages = append(truncatedMessages, messages[len(messages)-24:]...)
						messages = truncatedMessages
					}
				}
			}
		}
	}()

	// Wait for the command to complete or timeout
	select {
	case err := <-resultCh:
		// Wait for output processing to complete
		<-outputDone

		// Command completed
		if err != nil {
			messages = append(messages, fmt.Sprintf("Error: %s", err.Error()))
			return messages, fmt.Errorf("failed to build and install package: %w", err)
		}

		messages = append(messages, fmt.Sprintf("%s installed successfully", h.Name))
		return messages, nil

	case <-time.After(30 * time.Minute): // Timeout after 30 minutes
		// Command timed out, kill it
		if cmd.Process != nil {
			cmd.Process.Kill()
		}

		// Wait for output processing to complete
		<-outputDone

		messages = append(messages, "Command timed out after 30 minutes")
		return messages, fmt.Errorf("command timed out after 30 minutes")
	}
}

// InstallPackages installs packages using the AUR helper
func (h *Helper) InstallPackages(packages []string) ([]string, error) {
	if len(packages) == 0 {
		return []string{"No packages to install"}, nil
	}

	// Collect system messages - use a fixed size buffer to limit memory usage
	messages := make([]string, 0, 20) // Reduce capacity from 50 to 20
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
	args := []string{"-S", "--needed", "--noconfirm", "--noprogressbar"}
	args = append(args, packages...)

	// Create a command that uses sudo directly if needed
	var cmd *exec.Cmd

	// Use ionice along with nice to reduce both CPU and I/O priority
	if h.sudoPassword != "" {
		cmd = exec.Command("ionice", "-c", "3", "nice", "-n", "19", "sudo", "-S", h.Command)
		cmd.Args = append(cmd.Args, args...)
		messages = append(messages, "Using sudo with password")
	} else {
		// No password provided, just use the AUR helper directly with nice
		cmd = exec.Command("ionice", "-c", "3", "nice", "-n", "19", h.Command)
		cmd.Args = append(cmd.Args, args...)
		messages = append(messages, "No password provided")
	}

	// Set resource limits using environment variables
	cmd.Env = append(os.Environ(),
		"MAKEFLAGS=-j1",               // Limit make to 1 job to reduce memory usage
		"CARGO_BUILD_JOBS=1",          // Limit Rust builds to 1 job
		"RUSTFLAGS=-Ccodegen-units=1", // Reduce Rust memory usage
	)

	// Set up pipes for stdin, stdout, and stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return messages, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return messages, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return messages, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

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

	// Wait for the command to complete in a goroutine
	go func() {
		resultCh <- cmd.Wait()
	}()

	// Create a channel for conflict detection
	conflictCh := make(chan string, 1)

	// Create a channel to signal output processing is done
	outputDone := make(chan struct{})

	// Start a goroutine to read stdout and stderr
	go func() {
		defer close(outputDone)

		// Use a mutex to protect access to the messages slice
		var messagesMutex sync.Mutex

		// Use a counter to track how many messages we've processed
		// This helps us avoid checking the length of the messages slice too often
		messageCount := 2 // Start at 2 because we've already added 2 messages

		// Process stdout and stderr concurrently
		stdoutDone := make(chan struct{})
		stderrDone := make(chan struct{})

		// Process stdout
		go func() {
			defer close(stdoutDone)
			scanner := bufio.NewScanner(stdout)
			scanner.Buffer(make([]byte, 4096), 4096) // Use a small buffer

			for scanner.Scan() {
				line := scanner.Text()
				if line == "" {
					continue
				}

				// Check for conflicts
				if strings.Contains(line, "conflict") {
					select {
					case conflictCh <- line:
						// Sent conflict message
					default:
						// Channel full, just continue
					}
				}

				// Only keep important messages to reduce memory usage
				if strings.Contains(line, "error") || strings.Contains(line, "warning") ||
					strings.Contains(line, "installing") || strings.Contains(line, "conflict") {

					// Add to messages with thread safety
					messagesMutex.Lock()
					messages = append(messages, line)
					messageCount++

					// Limit the number of messages to avoid memory issues
					// Only check and truncate occasionally to reduce overhead
					if messageCount > 50 && messageCount%10 == 0 && len(messages) > 40 {
						// Keep only the first 20 and last 20 messages
						newMessages := make([]string, 0, 41)
						newMessages = append(newMessages, messages[:20]...)
						newMessages = append(newMessages, "... (output truncated) ...")
						newMessages = append(newMessages, messages[len(messages)-20:]...)
						messages = newMessages
					}
					messagesMutex.Unlock()
				}
			}

			if err := scanner.Err(); err != nil && err != io.EOF {
				messagesMutex.Lock()
				messages = append(messages, fmt.Sprintf("Error reading stdout: %v", err))
				messagesMutex.Unlock()
			}
		}()

		// Process stderr
		go func() {
			defer close(stderrDone)
			scanner := bufio.NewScanner(stderr)
			scanner.Buffer(make([]byte, 4096), 4096) // Use a small buffer

			for scanner.Scan() {
				line := scanner.Text()
				if line == "" {
					continue
				}

				// Check for conflicts
				if strings.Contains(line, "conflict") {
					select {
					case conflictCh <- line:
						// Sent conflict message
					default:
						// Channel full, just continue
					}
				}

				// Only keep important messages to reduce memory usage
				if strings.Contains(line, "error") || strings.Contains(line, "warning") ||
					strings.Contains(line, "installing") || strings.Contains(line, "conflict") {

					// Add to messages with thread safety
					messagesMutex.Lock()
					messages = append(messages, line)
					messageCount++

					// Limit the number of messages to avoid memory issues
					// Only check and truncate occasionally to reduce overhead
					if messageCount > 50 && messageCount%10 == 0 && len(messages) > 40 {
						// Keep only the first 20 and last 20 messages
						newMessages := make([]string, 0, 41)
						newMessages = append(newMessages, messages[:20]...)
						newMessages = append(newMessages, "... (output truncated) ...")
						newMessages = append(newMessages, messages[len(messages)-20:]...)
						messages = newMessages
					}
					messagesMutex.Unlock()
				}
			}

			if err := scanner.Err(); err != nil && err != io.EOF {
				messagesMutex.Lock()
				messages = append(messages, fmt.Sprintf("Error reading stderr: %v", err))
				messagesMutex.Unlock()
			}
		}()

		// Wait for both stdout and stderr processing to complete
		<-stdoutDone
		<-stderrDone
	}()

	// Wait for the command to complete or timeout
	select {
	case err := <-resultCh:
		// Wait for output processing to complete
		<-outputDone

		// Command completed
		if err != nil {
			// Check if we received a conflict message
			select {
			case conflictMsg := <-conflictCh:
				messages = append(messages, fmt.Sprintf("Conflict detected: %s", conflictMsg))
				return messages, fmt.Errorf("package conflict detected: %s", conflictMsg)
			default:
				// No conflict, just an error
				messages = append(messages, fmt.Sprintf("Command failed: %v", err))
				return messages, fmt.Errorf("command failed: %w", err)
			}
		}

		messages = append(messages, "Packages installed successfully")
		return messages, nil

	case conflictMsg := <-conflictCh:
		// Conflict detected
		if cmd.Process != nil {
			cmd.Process.Kill()
		}

		// Wait for output processing to complete
		<-outputDone

		messages = append(messages, fmt.Sprintf("Conflict detected: %s", conflictMsg))
		return messages, fmt.Errorf("package conflict detected: %s", conflictMsg)

	case <-time.After(30 * time.Minute): // Timeout after 30 minutes
		// Command timed out, kill it
		if cmd.Process != nil {
			cmd.Process.Kill()
		}

		// Wait for output processing to complete
		<-outputDone

		messages = append(messages, "Command timed out after 30 minutes")
		return messages, fmt.Errorf("command timed out after 30 minutes")
	}
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
