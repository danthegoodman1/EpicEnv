package cmd

import (
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

// listInvitesCmd represents the list-invites command
var listInvitesCmd = &cobra.Command{
	Use:   "list-invites",
	Short: "List all GitHub users invited to the environment",
	Long: `List all GitHub users that have been invited to the environment.
	
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
		logger.Info().Msgf("No users are invited to environment '%s'", env)
		return
	}

	// Extract unique usernames
	usernames := lo.Uniq(lo.Map(keysFile.EncryptedKeys, func(key EncryptedKey, _ int) string {
		return key.Username
	}))

	// Get count of keys per user
	userKeyCounts := make(map[string]int)
	for _, key := range keysFile.EncryptedKeys {
		userKeyCounts[key.Username]++
	}

	logger.Info().Msgf("Users invited to environment '%s':", env)
	for _, username := range usernames {
		logger.Info().Msgf("- %s (%d keys)", username, userKeyCounts[username])
	}
}
