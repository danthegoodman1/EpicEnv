/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"fmt"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"os"
)

// setCmd represents the set command
var setCmd = &cobra.Command{
	Use:   "set KEY [VALUE]",
	Short: "Set an environment variable",
	Long: `Set an environment variable

Use -e to set the environment, can omit to use the current.

Use -p to set a personal variable.

Omit [VALUE] to collect from stdin

If you attempt to normal set a personal variable, it will update the personal variable instead. To make a personal variable shared, first rm the variable, then set it again as shared.`,
	Run:        runSet,
	Args:       cobra.RangeArgs(1, 2),
	ArgAliases: []string{"env", "key", "value"},
}

func init() {
	rootCmd.AddCommand(setCmd)

	setCmd.Flags().BoolP("personal", "p", false, "Set this as a personal environment if it doesn't exist")
}

func runSet(cmd *cobra.Command, args []string) {
	key := args[0]
	val := ""
	if len(args) == 1 {
		val = readStdinHidden(fmt.Sprintf("%s> ", key))
	} else {
		val = args[1]
	}

	env := getEnvOrFlag(cmd)

	personal := false
	if cmd.Flag("personal") != nil {
		personal = cmd.Flag("personal").Value.String() == "true"
	}
	setEnvVar(env, key, val, personal)

	logger.Info().Msgf("Updated %s", key)

	if os.Getenv("EPICENV") != "" {
		logger.Info().Msgf("To reload the current environment, run:\n\tsource .epicenv/%s/activate", env)
	}
}

func setEnvVar(env, key, val string, personal bool) {
	envMap := loadEnv(env)
	// Check if we are setting a personal env var

	if envVar, exists := envMap[key]; exists && personal && !envVar.Personal {
		logger.Fatal().Msgf("Attempting to set an existing shared env var \"%s\" as personal, please rm this env var and set again", key)
	}
	if envVar, exists := envMap[key]; exists && !personal && envVar.Personal {
		// update and warn
		personal = envVar.Personal
		logger.Warn().Msg("Attempting to set an existing personal env var as shared, will set as personal")
	}

	secretsFile, err := readSecretsFile(env, personal)
	if errors.Is(err, os.ErrNotExist) {
		secretsFile = &SecretsFile{}
	} else if err != nil {
		logger.Fatal().Err(err).Msg("error reading secrets file")
	}

	// Encrypt the value
	symKey, err := loadSymmetricKey(env)
	if err != nil {
		logger.Fatal().Err(err).Msg("error loading symmetric key")
	}

	encrypted, err := encryptAESGCM(symKey, val)
	if err != nil {
		logger.Fatal().Err(err).Msg("error encrypting value")
	}

	if _, exists := envMap[key]; exists {
		// Replace the value
		logger.Debug().Msgf("Var %s exists", key)
		_, idx, _ := lo.FindIndexOf(secretsFile.Secrets, func(item EncryptedSecret) bool {
			return item.Name == key
		})

		if idx == -1 {
			logger.Fatal().Msg("unable to find secret again... this is a bug, please report")
		}

		secretsFile.Secrets[idx].Value = encrypted
	} else {
		logger.Debug().Msgf("Var %s does not exist", key)
		// Append the value
		secretsFile.Secrets = append(secretsFile.Secrets, EncryptedSecret{
			Name:     key,
			Personal: personal,
			Value:    encrypted,
		})

		if personal {
			// We need to mark it in the shared secrets that it exists now
			sharedSecrets, err := readSecretsFile(env, false)
			if errors.Is(err, os.ErrNotExist) {
				sharedSecrets = &SecretsFile{}
			} else if err != nil {
				logger.Fatal().Err(err).Msg("error reading shared secrets file")
			}

			sharedSecrets.Secrets = append(sharedSecrets.Secrets, EncryptedSecret{
				Name:     key,
				Personal: true,
				Value:    "",
			})

			err = writeSecretsFile(env, *sharedSecrets, false)
			if err != nil {
				logger.Fatal().Err(err).Msg("error writing shared secrets file")
			}
		}
	}

	err = writeSecretsFile(env, *secretsFile, personal)
	if err != nil {
		logger.Fatal().Err(err).Msg("error writing secrets file")
	}
}
