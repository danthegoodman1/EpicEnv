package cmd

import (
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

// listInvitesCmd represents the list-invites command
var listInvitesCmd = &cobra.Command{
	Use:   "list-invites",
	Short: "List all GitHub users and headless keys invited to the environment",
	Long: `List all GitHub users and headless keys that have been invited to the environment.

Example:
  epicenv list-invites -e prod`,
	Run: runListInvites,
}

func init() {
	rootCmd.AddCommand(listInvitesCmd)
}

func runListInvites(cmd *cobra.Command, args []string) {
	env := getEnvOrFlag(cmd)

	// Load the keys file
	keysFile, err := readKeysFile(env)
	if err != nil {
		logger.Fatal().Err(err).Msg("error reading keys file")
	}

	if len(keysFile.EncryptedKeys) == 0 {
		logger.Info().Msgf("No users or keys are invited to environment '%s'", env)
		return
	}

	// Extract unique usernames
	usernames := lo.Uniq(lo.Map(keysFile.EncryptedKeys, func(key EncryptedKey, _ int) string {
		return key.Username
	}))

	// Get count of keys per user and track which are headless
	userKeyCounts := make(map[string]int)
	userIsHeadless := make(map[string]bool)

	for _, key := range keysFile.EncryptedKeys {
		userKeyCounts[key.Username]++
		if key.IsHeadless {
			userIsHeadless[key.Username] = true
		}
	}

	// Separate GitHub users from headless keys
	var githubUsers []string
	var headlessKeys []string

	for _, username := range usernames {
		if userIsHeadless[username] {
			headlessKeys = append(headlessKeys, username)
		} else {
			githubUsers = append(githubUsers, username)
		}
	}

	logger.Info().Msgf("Keys invited to environment '%s':", env)

	if len(githubUsers) > 0 {
		logger.Info().Msg("GitHub Users:")
		for _, username := range githubUsers {
			logger.Info().Msgf("- %s (%d keys)", username, userKeyCounts[username])
		}
	}

	if len(headlessKeys) > 0 {
		logger.Info().Msg("Headless Keys:")
		for _, keyname := range headlessKeys {
			logger.Info().Msgf("- %s (%d keys)", keyname, userKeyCounts[keyname])
		}
	}
}
