package cmd

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"syscall"

	"golang.org/x/term"

	"github.com/danthegoodman1/epicenv/gologger"
)

var (
	logger = gologger.NewLogger()
	// Cache the epicenv directory path
	epicEnvDir string
)

func readStdinHidden(prompt string) string {
	fmt.Print(prompt)
	// IDE might complain, but the cast is necessary for some OSs, because Stdin is a var instead of an untyped const
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		panic(err)
	}
	fmt.Println() // Move to the next line after input
	return string(bytePassword)
}

var ErrEnvDirNotFound = errors.New("epicenv directory not found")

// findEpicEnvDir walks up the directory tree until it either finds a .epicenv directory,
// hits the filesystem root, or encounters a permission issue
func findEpicEnvDir() (string, error) {
	// Return cached path if we've already found it
	if epicEnvDir != "" {
		return epicEnvDir, nil
	}

	// Start from the current directory
	currentDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("error getting current directory: %w", err)
	}

	// Walk up the directory tree
	for {
		// Check if .epicenv exists in the current directory
		potentialEpicEnvDir := filepath.Join(currentDir, ".epicenv")
		_, err := os.Stat(potentialEpicEnvDir)
		if err == nil {
			// Found it!
			epicEnvDir = currentDir
			return currentDir, nil
		}

		// Check if it's a permission error
		if !os.IsNotExist(err) {
			return "", fmt.Errorf("error checking for .epicenv in %s: %w", currentDir, err)
		}

		// Go up one directory
		parentDir := filepath.Dir(currentDir)

		// If we're at the root, we can't go up any further
		if parentDir == currentDir {
			break
		}

		currentDir = parentDir
	}

	// If we got here, we didn't find the .epicenv directory
	return "", ErrEnvDirNotFound
}

// getEpicEnvPath returns the path to the .epicenv directory
func getEpicEnvPath() string {
	dir, err := findEpicEnvDir()
	if err != nil {
		// Default to current directory if not found
		return ".epicenv"
	}
	return filepath.Join(dir, ".epicenv")
}

// Optimistic check, does not check for permissions
func envExists(env string) bool {
	epicEnvPath := getEpicEnvPath()
	if _, err := os.Stat(path.Join(epicEnvPath, env)); errors.Is(err, os.ErrNotExist) {
		return false
	} else if err != nil {
		logger.Fatal().Err(err).Msg("error checking if file exists")
	}

	return true
}

// listEnvironments returns a list of all available environments
func listEnvironments() ([]string, error) {
	epicEnvPath := getEpicEnvPath()

	// Check if .epicenv directory exists
	if _, err := os.Stat(epicEnvPath); errors.Is(err, os.ErrNotExist) {
		return []string{}, nil
	} else if err != nil {
		return nil, fmt.Errorf("error checking .epicenv directory: %w", err)
	}

	// Read directory contents
	entries, err := os.ReadDir(epicEnvPath)
	if err != nil {
		return nil, fmt.Errorf("error reading .epicenv directory: %w", err)
	}

	var environments []string
	for _, entry := range entries {
		// Only include directories that are not hidden
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") && !strings.HasPrefix(entry.Name(), "temp") {
			environments = append(environments, entry.Name())
		}
	}

	return environments, nil
}
