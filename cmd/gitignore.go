package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

func prepareGitIgnore() error {
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

	fileString := string(fileContent)

	fmt.Println(fileString)

	ignoreStat, err := os.Stat(".gitignore")
	if err != nil {
		logger.Fatal().Err(err).Msg("error getting stats on .gitignore")
	}

	if !strings.Contains(fileString, ".epicenv/*/personal_keys.json") {
		fileString += "\n.epicenv/*/personal_keys.json\n"
	}
	if !strings.Contains(fileString, ".epicenv/temp*") {
		fileString += "\n.epicenv/temp*\n"
	}

	err = os.WriteFile(".gitignore", []byte(fileString), ignoreStat.Mode())
	if err != nil {
		return fmt.Errorf("error in WriteFile for .gitignore: %w", err)
	}

	return nil
}
