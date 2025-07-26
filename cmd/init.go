package cmd

import (
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize EpicEnv",

	Run: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) {
	// name env (default)
	env := readStdin("Create a new environment [default]> ")
	if env == "" {
		env = "local"
	}
	logger.Debug().Msgf("Got env %s", env)

	// check if env already exists
	if envExists(env) {
		logger.Fatal().Msgf("Environment %s already exists", env)
	}

	logger.Debug().Msgf("Env %s does not exist, creating", env)

	// collect their keys from github, verify they exist in $HOME/.ssh/
	githubUser := readStdin("What is your GitHub username? So I can fetch your public key(s)> ")
	if githubUser == "" {
		logger.Fatal().Msg("I need your GitHub username...")
	}

	logger.Debug().Msgf("Got github username %s", githubUser)

	pubKeys, err := getKeysForGithubUsername(githubUser)
	if err != nil {
		logger.Fatal().Err(err).Msgf("error getting keys from github for %s", githubUser)
	}

	logger.Debug().Msgf("Got %d keys from github for %s", len(pubKeys), githubUser)

	if len(pubKeys) == 0 {
		logger.Fatal().Msgf("No keys found for GitHub user %s, please add an SSH key to set up EpicEnv!", githubUser)
	}

	foundKeys := findPrivateKeysForPublicKeys(pubKeys)
	logger.Debug().Msgf("Found %d private keys in $HOME/.ssh", len(foundKeys))
	if len(foundKeys) == 0 {
		logger.Fatal().Msgf("Did not find any of the keys in GitHub for %s in $HOME/.ssh/", githubUser)
	}

	// append personal secrets to gitignore or create it
	err = prepareGitIgnore()
	if err != nil {
		logger.Fatal().Err(err).Msg("error creating gitignore")
	}

	// create initial symmetric key
	aesKey := generateAESKey()

	// write keys to disk
	keysFile := KeysFile{
		EncryptedKeys: lo.Map(foundKeys, func(item keyPair, index int) EncryptedKey {
			// Encrypt the symmetric key with their private key
			encryptedAESKey, err := encryptWithPublicKey(aesKey, item.publicKeyContent)
			if err != nil {
				logger.Fatal().Err(err).Msg("error in encryptWithPublicKey")
			}

			return EncryptedKey{
				Username:           githubUser,
				PublicKey:          item.publicKeyContent,
				EncryptedSharedKey: encryptedAESKey,
			}
		}),
	}
	err = writeKeysFile(env, keysFile)
	if err != nil {
		logger.Fatal().Err(err).Msg("error writing keys file")
	}

	err = generateActivateSource(env)
	if err != nil {
		logger.Fatal().Err(err).Msg("erorr generating activate source")
	}

	logger.Info().Msgf("Initialized %s", env)
}
