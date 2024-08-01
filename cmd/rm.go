/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"os"
)

// rmCmd represents the rm command
var rmCmd = &cobra.Command{
	Use:   "rm",
	Short: "Remove a variable from the environment",
	Run:   runRm,
	Args:  cobra.ExactArgs(1),
}

func init() {
	rootCmd.AddCommand(rmCmd)
}

func runRm(cmd *cobra.Command, args []string) {
	key := args[0]
	env := getEnvOrFlag(cmd)
	envMap := loadEnv(env)

	envVar, exists := envMap[key]
	if !exists {
		logger.Warn().Msgf("The environment variable %s doesn't exist!", key)
		os.Exit(1)
	}

	// Load the shared secrets
	secretsFile, err := readSecretsFile(env, false)
	if errors.Is(err, os.ErrNotExist) {
		logger.Fatal().Msg("No secrets file found to delete from")
	} else if err != nil {
		logger.Fatal().Err(err).Msg("error reading secrets file")
	}

	// Purge it
	secretsFile.Secrets = lo.Filter(secretsFile.Secrets, func(item EncryptedSecret, index int) bool {
		return item.Name != key
	})
	err = writeSecretsFile(env, *secretsFile, false)
	if err != nil {
		logger.Fatal().Err(err).Msg("error writing updated secrets file")
	}

	// If personal, remove it there too
	if envVar.Personal {
		logger.Debug().Msgf("%s is personal, removing from personal secrets", key)

		secretsFile, err = readSecretsFile(env, false)
		if errors.Is(err, os.ErrNotExist) {
			logger.Fatal().Msg("No secrets file found to delete from")
		} else if err != nil {
			logger.Fatal().Err(err).Msg("error reading personal secrets file")
		}

		// Purge it
		secretsFile.Secrets = lo.Filter(secretsFile.Secrets, func(item EncryptedSecret, index int) bool {
			return item.Name != key
		})
		err = writeSecretsFile(env, *secretsFile, true)
		if err != nil {
			logger.Fatal().Err(err).Msg("error writing updated personal secrets file")
		}
	}

	logger.Info().Msgf("Removed %s", key)
}
