package cmd

import (
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"os"
)

// inviteCmd represents the invite command
var inviteCmd = &cobra.Command{
	Use:   "invite",
	Short: "Invite a GitHub user to the EpicEnv, allowing them to decrypt the environment",
	Run:   runInvite,
	Args:  cobra.ExactArgs(1),
}

func init() {
	rootCmd.AddCommand(inviteCmd)
}

func runInvite(cmd *cobra.Command, args []string) {
	githubUsername := args[0]
	env := getEnvOrFlag(cmd)

	// get their key(s)
	foundKeys, err := getKeysForGithubUsername(githubUsername)
	if err != nil {
		logger.Fatal().Err(err).Msgf("error getting keys from github for %s", githubUsername)
	}

	logger.Debug().Msgf("Got %d keys from github for %s", len(foundKeys), githubUsername)

	if len(foundKeys) == 0 {
		logger.Fatal().Msgf("No keys found for GitHub user %s, please add an SSH key to set up EpicEnv!", githubUsername)
	}

	// encrypt the sym key with their key
	keysFile, err := readKeysFile(env)
	if err != nil {
		logger.Fatal().Err(err).Msg("error loading keys file")
	}

	symKey, err := loadSymmetricKey(env)
	if err != nil {
		logger.Fatal().Err(err).Msg("error loading symmetric key")
	}

	added := 0
	for _, key := range foundKeys {
		if lo.ContainsBy(keysFile.EncryptedKeys, func(item EncryptedKey) bool {
			return item.PublicKey == key
		}) {
			// If it already exists, continue
			logger.Debug().Msgf("skipping existing key like %s", key[:16])
			continue
		}

		// Encrypt the sym key with their pub key
		encSymKey, err := encryptWithPublicKey(symKey, key)
		if err != nil {
			logger.Fatal().Err(err).Msg("error encrypting with public key")
		}
		encKey := EncryptedKey{
			Username:           githubUsername,
			PublicKey:          key,
			EncryptedSharedKey: encSymKey,
		}
		keysFile.EncryptedKeys = append(keysFile.EncryptedKeys, encKey)

		added++
	}

	if added == 0 {
		logger.Warn().Msg("No new keys added")
		os.Exit(0)
	}

	// Write the file
	err = writeKeysFile(env, *keysFile)
	if err != nil {
		logger.Fatal().Err(err).Msg("error writing keys file")
	}
	logger.Info().Msgf("Added GitHub user %s's keys", githubUsername)
}
