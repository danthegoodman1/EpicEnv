/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

// uninviteCmd represents the uninvite command
var uninviteCmd = &cobra.Command{
	Use:   "uninvite [name]",
	Short: "Uninvite a GitHub user or remove a headless key from the environment",
	Long: `Uninvite a GitHub user or remove a headless key from the environment, removing their access by re-encrypting.
This is not a replacement for rotating secrets!

Examples:
  epicenv uninvite username    # Uninvite a GitHub user
  epicenv uninvite keyname     # Remove a headless key`,
	Run:  runUninvite,
	Args: cobra.ExactArgs(1),
}

func init() {
	rootCmd.AddCommand(uninviteCmd)
}

func runUninvite(cmd *cobra.Command, args []string) {
	name := args[0]
	env := getEnvOrFlag(cmd)

	// Load in the keys
	keysFile, err := readKeysFile(env)
	if err != nil {
		logger.Fatal().Err(err).Msg("error reading keys file")
	}

	// Check if the user or key exists
	if !lo.ContainsBy(keysFile.EncryptedKeys, func(item EncryptedKey) bool {
		return item.Username == name
	}) {
		logger.Fatal().Msgf("User or key '%s' is not invited to this environment", name)
	}

	// Load the symmetric key (so we know that we are invited to the env)
	_, err = loadSymmetricKey(env)
	if err != nil {
		logger.Fatal().Err(err).Msg("error loading symmetric key")
	}

	// Check if it's a headless key or GitHub user
	isHeadless := false
	for _, key := range keysFile.EncryptedKeys {
		if key.Username == name && key.IsHeadless {
			isHeadless = true
			break
		}
	}

	// Remove the key
	keysFile.EncryptedKeys = lo.Filter(keysFile.EncryptedKeys, func(item EncryptedKey, index int) bool {
		return item.Username != name
	})

	// Write the updated keys file
	err = writeKeysFile(env, *keysFile)
	if err != nil {
		logger.Fatal().Err(err).Msg("error writing keys file")
	}

	if isHeadless {
		logger.Info().Msgf("Removed headless key '%s' **THIS IS NOT A REPLACEMENT FOR ROTATING SECRETS!**", name)
	} else {
		logger.Info().Msgf("Removed GitHub user '%s' **THIS IS NOT A REPLACEMENT FOR ROTATING SECRETS!**", name)
	}
}
