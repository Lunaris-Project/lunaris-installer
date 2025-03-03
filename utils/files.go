package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// CopyDir recursively copies a directory tree, attempting to preserve permissions.
func CopyDir(src, dst string) error {
	// Get properties of source dir
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("error getting stats for source directory: %w", err)
	}

	// Create destination dir
	if err = os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return fmt.Errorf("error creating destination directory: %w", err)
	}

	// Read source directory
	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("error reading source directory: %w", err)
	}

	// Copy each entry
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			// Recursively copy the directory
			if err = CopyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// Copy the file
			if err = CopyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// CopyFile copies a single file from src to dst
func CopyFile(src, dst string) error {
	// Open the source file
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("error opening source file: %w", err)
	}
	defer srcFile.Close()

	// Get source file info
	srcInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("error getting source file info: %w", err)
	}

	// Create the destination file
	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return fmt.Errorf("error creating destination file: %w", err)
	}
	defer dstFile.Close()

	// Copy the contents
	if _, err = io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("error copying file contents: %w", err)
	}

	return nil
}

// GetHomeDir returns the user's home directory
func GetHomeDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("error getting user home directory: %w", err)
	}
	return home, nil
}

// CopyConfigDirs copies configuration directories to the user's home directory
func CopyConfigDirs(repoPath string, configDirs []string) error {
	homeDir, err := GetHomeDir()
	if err != nil {
		return err
	}

	for _, dir := range configDirs {
		srcPath := filepath.Join(repoPath, dir)
		dstPath := filepath.Join(homeDir, dir)

		// Check if the source directory exists
		if _, err := os.Stat(srcPath); os.IsNotExist(err) {
			fmt.Printf("Warning: Source directory %s does not exist, skipping\n", srcPath)
			continue
		}

		// Check if the destination directory exists
		if _, err := os.Stat(dstPath); err == nil {
			// Create a backup of the existing directory
			backupPath := dstPath + ".bak"
			fmt.Printf("Backing up existing directory %s to %s\n", dstPath, backupPath)

			// Remove any existing backup
			os.RemoveAll(backupPath)

			// Rename the existing directory to the backup path
			if err := os.Rename(dstPath, backupPath); err != nil {
				return fmt.Errorf("error backing up directory %s: %w", dstPath, err)
			}
		}

		// Copy the directory
		fmt.Printf("Copying %s to %s\n", srcPath, dstPath)
		if err := CopyDir(srcPath, dstPath); err != nil {
			return fmt.Errorf("error copying directory %s: %w", srcPath, err)
		}
	}

	return nil
}

// GetExecutablePath returns the path of the current executable
func GetExecutablePath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("error getting executable path: %w", err)
	}
	return filepath.Dir(exe), nil
}

// GetRepoPath returns the path of the repository
func GetRepoPath() (string, error) {
	// Get the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("error getting current working directory: %w", err)
	}

	// Check if we're in the repository
	if _, err := os.Stat(filepath.Join(cwd, ".git")); err == nil {
		return cwd, nil
	}

	// Check if we're in a subdirectory of the repository
	for dir := cwd; dir != "/"; dir = filepath.Dir(dir) {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir, nil
		}
	}

	// If we can't find the repository, use the current directory
	return cwd, nil
}
