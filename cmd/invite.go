package cmd

import (
	"os"
	"strings"

	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

// inviteCmd represents the invite command
var inviteCmd = &cobra.Command{
	Use:   "invite [name]",
	Short: "Invite a GitHub user or add a headless key to the EpicEnv",
	Long: `Invite a GitHub user to the EpicEnv, allowing them to decrypt the environment,
or add a headless key (not associated with a GitHub user).

Examples:
  epicenv invite username               # Invite GitHub user by username
  epicenv invite keyname --path key.pub # Add a headless key from a file`,
	Run:  runInvite,
	Args: cobra.ExactArgs(1),
}

var pathFlag string

func init() {
	rootCmd.AddCommand(inviteCmd)
	inviteCmd.Flags().StringVar(&pathFlag, "path", "", "Path to public key file (for headless keys)")
}

func runInvite(cmd *cobra.Command, args []string) {
	name := args[0]
	env := getEnvOrFlag(cmd)

	// Check if using path flag for headless key
	usingPath := pathFlag != ""

	// Load the keys file
	keysFile, err := readKeysFile(env)
	if err != nil {
		logger.Fatal().Err(err).Msg("error loading keys file")
	}

	// Check if the name is already in use
	existingKey := lo.ContainsBy(keysFile.EncryptedKeys, func(key EncryptedKey) bool {
		return key.Username == name
	})

	if existingKey {
		// Name already exists
		logger.Fatal().Msgf("The name '%s' is already in use. Please use a different name.", name)
	}

	symKey, err := loadSymmetricKey(env)
	if err != nil {
		logger.Fatal().Err(err).Msg("error loading symmetric key")
	}

	var foundKeys []string

	if usingPath {
		// Handle headless key from file
		keyData, err := os.ReadFile(pathFlag)
		if err != nil {
			logger.Fatal().Err(err).Msgf("error reading key file %s", pathFlag)
		}

		publicKey := strings.TrimSpace(string(keyData))
		foundKeys = []string{publicKey}
	} else {
		// Handle GitHub user
		foundKeys, err = getKeysForGithubUsername(name)
		if err != nil {
			logger.Fatal().Err(err).Msgf("error getting keys from github for %s", name)
		}

		logger.Debug().Msgf("Got %d keys from github for %s", len(foundKeys), name)

		if len(foundKeys) == 0 {
			logger.Fatal().Msgf("No keys found for GitHub user %s, please add an SSH key to set up EpicEnv!", name)
		}
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
			Username:           name,
			PublicKey:          key,
			EncryptedSharedKey: encSymKey,
			IsHeadless:         usingPath,
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

	if usingPath {
		logger.Info().Msgf("Added headless key '%s'", name)
	} else {
		logger.Info().Msgf("Added GitHub user %s's keys", name)
	}
}
