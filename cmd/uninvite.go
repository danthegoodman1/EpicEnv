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
	Use:   "uninvite",
	Short: "Uninvite a GitHub collaborator from the environment, removing their access by re-encrypting. This is not a replacement for rotating secrets!",
	Run:   runUninvite,
	Args:  cobra.ExactArgs(1),
}

func init() {
	rootCmd.AddCommand(uninviteCmd)
}

func runUninvite(cmd *cobra.Command, args []string) {
	githubUser := args[0]
	env := getEnvOrFlag(cmd)

	// Load in the keys
	keysFile, err := readKeysFile(env)
	if err != nil {
		logger.Fatal().Err(err).Msg("error reading keys file")
	}

	// Check if the user exists
	if !lo.ContainsBy(keysFile.EncryptedKeys, func(item EncryptedKey) bool {
		return item.Username == githubUser
	}) {
		logger.Fatal().Msgf("GitHub user %s is not invited to this environment", githubUser)
	}

	// Load the symmetric key (so we know that we are invited to the env)
	_, err = loadSymmetricKey(env)
	if err != nil {
		logger.Fatal().Err(err).Msg("error loading symmetric key")
	}

	// Remove the key
	keysFile.EncryptedKeys = lo.Filter(keysFile.EncryptedKeys, func(item EncryptedKey, index int) bool {
		return item.Username != githubUser
	})

	// Write the updated keys file
	err = writeKeysFile(env, *keysFile)
	if err != nil {
		logger.Fatal().Err(err).Msg("error writing keys file")
	}

	logger.Info().Msgf("Removed %s **THIS IS NOT A REPLACEMENT FOR ROTATING SECRETS!**", githubUser)
}
