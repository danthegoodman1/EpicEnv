package cmd

import (
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init GITHUB_USERNAME",
	Short: "Initialize EpicEnv",
	Long: `Initialize EpicEnv with a new environment.

Provide your GitHub username as an argument to fetch your public SSH keys.
Use the -e flag to specify the environment name (defaults to "local").

Examples:
  epicenv init danthegoodman1                  # Creates default "local" environment
  epicenv init danthegoodman1 -e staging       # Creates "staging" environment`,
	Run:  runInit,
	Args: cobra.ExactArgs(1),
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) {
	// Get GitHub username from positional argument
	githubUser := args[0]
	logger.Debug().Msgf("Got github username %s", githubUser)

	// Get environment name from flag or use default
	env, err := cmd.Flags().GetString("environment")
	if err != nil {
		logger.Fatal().Err(err).Msg("error getting environment flag")
	}
	if env == "" {
		env = "local"
	}
	logger.Debug().Msgf("Got env %s", env)

	// check if env already exists
	if envExists(env) {
		logger.Fatal().Msgf("Environment %s already exists", env)
	}

	logger.Debug().Msgf("Env %s does not exist, creating", env)

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
