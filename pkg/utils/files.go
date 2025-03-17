package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

// CopyFile copies a file from src to dst
func CopyFile(src, dst string) error {
	// Open the source file
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	// Create the destination file
	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	// Copy the contents
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy file contents: %w", err)
	}

	// Get the file mode of the source file
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to get source file info: %w", err)
	}

	// Set the file mode of the destination file
	if err := os.Chmod(dst, srcInfo.Mode()); err != nil {
		return fmt.Errorf("failed to set file mode: %w", err)
	}

	return nil
}

// CopyDir copies a directory from src to dst
func CopyDir(src, dst string) error {
	// Get the file info of the source directory
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to get source directory info: %w", err)
	}

	// Create the destination directory
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Read the source directory
	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("failed to read source directory: %w", err)
	}

	// Copy each entry
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			// Recursively copy the directory
			if err := CopyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// Copy the file
			if err := CopyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// CopyDirWithLowMemory copies a directory from src to dst with low memory usage
// It uses a worker pool pattern to limit concurrent operations and reduce memory usage
func CopyDirWithLowMemory(src, dst string) error {
	// Get the file info of the source directory
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to get source directory info: %w", err)
	}

	// Create the destination directory
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Create a worker pool to limit concurrent operations
	// Use half the available CPUs to avoid excessive memory usage
	numWorkers := max(1, runtime.NumCPU()/2)

	// Create a channel for tasks
	tasks := make(chan copyTask, 100)

	// Create a WaitGroup to wait for all workers to finish
	var wg sync.WaitGroup

	// Create a channel to collect errors
	errCh := make(chan error, numWorkers)

	// Start workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range tasks {
				if task.isDir {
					// Create the directory
					if err := os.MkdirAll(task.dst, task.mode); err != nil {
						errCh <- fmt.Errorf("failed to create directory %s: %w", task.dst, err)
						return
					}
				} else {
					// Copy the file with a small buffer to reduce memory usage
					if err := copyFileWithSmallBuffer(task.src, task.dst, task.mode); err != nil {
						errCh <- fmt.Errorf("failed to copy file %s: %w", task.src, err)
						return
					}
				}
			}
		}()
	}

	// Walk the source directory and send tasks to workers
	err = filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate the destination path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		dstPath := filepath.Join(dst, relPath)

		// Skip the root directory as we've already created it
		if path == src {
			return nil
		}

		// Send a task to the worker pool
		tasks <- copyTask{
			src:   path,
			dst:   dstPath,
			isDir: info.IsDir(),
			mode:  info.Mode(),
		}

		return nil
	})

	// Close the tasks channel to signal workers to exit
	close(tasks)

	// Wait for all workers to finish
	wg.Wait()

	// Check if there were any errors
	select {
	case err := <-errCh:
		return err
	default:
		// No errors
	}

	return err
}

// copyTask represents a file or directory copy task
type copyTask struct {
	src   string
	dst   string
	isDir bool
	mode  os.FileMode
}

// copyFileWithSmallBuffer copies a file from src to dst using a small buffer
func copyFileWithSmallBuffer(src, dst string, mode os.FileMode) error {
	// Open the source file
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	// Create the destination file
	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	// Use a small buffer (4KB) to reduce memory usage
	buf := make([]byte, 4096)

	// Copy the file in chunks
	for {
		n, err := srcFile.Read(buf)
		if err != nil && err != io.EOF {
			return fmt.Errorf("failed to read from source file: %w", err)
		}

		if n == 0 {
			break
		}

		if _, err := dstFile.Write(buf[:n]); err != nil {
			return fmt.Errorf("failed to write to destination file: %w", err)
		}
	}

	// Set the file mode of the destination file
	if err := os.Chmod(dst, mode); err != nil {
		return fmt.Errorf("failed to set file mode: %w", err)
	}

	return nil
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
