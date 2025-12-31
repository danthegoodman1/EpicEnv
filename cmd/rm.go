/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"os"

	"github.com/samber/lo"
	"github.com/spf13/cobra"
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

	// Check if key exists in this environment's secrets
	keyInThisEnv := lo.ContainsBy(secretsFile.Secrets, func(item EncryptedSecret) bool {
		return item.Name == key
	})

	if !keyInThisEnv && !envVar.Personal {
		// Key must be coming from an underlay
		logger.Warn().Msgf("'%s' is defined in an underlay environment, not in '%s' - nothing to remove here", key, env)
		os.Exit(0)
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

		secretsFile, err = readSecretsFile(env, true)
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

	// Check if key will still be visible from underlay
	if isOverlay(env) {
		chain, err := getOverlayChain(env)
		if err == nil && len(chain) > 1 {
			// Check underlays (all except current env)
			for _, underlayEnv := range chain[:len(chain)-1] {
				underlaySecrets, err := readSecretsFile(underlayEnv, false)
				if err != nil {
					continue
				}
				if lo.ContainsBy(underlaySecrets.Secrets, func(item EncryptedSecret) bool {
					return item.Name == key
				}) {
					logger.Warn().Msgf("Note: '%s' still exists in underlay '%s' and will be visible", key, underlayEnv)
					break
				}
			}
		}
	}

	logger.Info().Msgf("Removed %s", key)
}
