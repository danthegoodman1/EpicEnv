package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

func addToGitIgnore() error {
	if _, err := os.Stat(".gitignore"); errors.Is(err, os.ErrNotExist) {
		// create the file
		logger.Debug().Msg("creating .gitignore")
		f, err := os.Create(".gitignore")
		if err != nil {
			return fmt.Errorf("error in os.Create for .gitignore: %w", err)
		}
		err = f.Close()
		if err != nil {
			return fmt.Errorf("error in f.Close(): %w", err)
		}
	} else if err != nil {
		logger.Fatal().Err(err).Msg("error checking if file exists")
	}

	// Check if the rows exist already
	fileContent, err := os.ReadFile(".gitignore")
	if err != nil {
		return fmt.Errorf("error in os.ReadFile: %w", err)
	}

	if strings.Contains(string(fileContent), ".epicenv/*/personal_keys.json") {
		logger.Debug().Msg("\".epicenv/*/personal_keys.json\" already in .gitignore, continuing")
		return nil
	}

	// Add personal keys
	f, err := os.OpenFile(".gitignore", os.O_WRONLY, 0777)
	if err != nil {
		return fmt.Errorf("error in os.Open: %w", err)
	}
	defer f.Close()
	_, err = f.WriteString(fmt.Sprintf("\n%s\n", ".epicenv/*/personal_keys.json"))
	if err != nil {
		return fmt.Errorf("error in WriteString: %w", err)
	}

	return nil
}
