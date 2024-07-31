package cmd

import (
	"errors"
	"fmt"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"os"
	"path"
	"strings"
	"time"
)

// internalGenCmd represents the internalGen command
var internalGenCmd = &cobra.Command{
	Use:   "zzz_INTERNAL_gen",
	Short: "FOR INTERNAL USE DO NOT RUN: Generates the temporary script to source",
	Run:   runInternalGenCmd,
}

func init() {
	rootCmd.AddCommand(internalGenCmd)
}

func runInternalGenCmd(cmd *cobra.Command, args []string) {
	env := cmd.Flag("environment").Value.String()
	logger.Debug().Msgf("running gen for env %s", env)

	keysFile, err := readKeysFile(env)
	if err != nil {
		logger.Fatal().Err(err).Msgf("error reading keys file, is it corrupted?")
	}

	keyPairs := findPrivateKeysForPublicKeys(lo.Map(keysFile.EncryptedKeys, func(item EncryptedKey, index int) string {
		return item.PublicKey
	}))

	if len(keyPairs) == 0 {
		logger.Fatal().Msgf("did not find any local private keys matching a known public key, are you invited to the %s environment?", env)
	}

	// Use the first one we found
	chosenKey := keyPairs[0]

	// decrypt symmetric key
	actualKey, found := lo.Find(keysFile.EncryptedKeys, func(item EncryptedKey) bool {
		return item.PublicKey == chosenKey.publicKeyContent
	})
	if !found {
		logger.Fatal().Msg("did not find the known public key again among the encrypted keys, this is a bug. Please report.")
	}

	symKey, err := decryptWithPrivateKey(actualKey.EncryptedSharedKey, chosenKey)
	if err != nil {
		logger.Fatal().Err(err).Msg("error decrypting symmetric key with private key")
	}

	// decrypt keys
	secretsFile, err := readSecretsFile(env, false)
	if errors.Is(err, os.ErrNotExist) {
		logger.Fatal().Msg("did not find a secrets file, you need to set some variables first!")
	}
	if err != nil {
		logger.Fatal().Err(err).Msg("error reading shared secrets file, is it corrupted?")
	}

	envVars := lo.Map(secretsFile.Secrets, func(item EncryptedSecret, index int) DecryptedSecret {
		envVar := DecryptedSecret{
			Name:     item.Name,
			Value:    "",
			Personal: false,
		}
		if !item.Personal {
			envVar.Value, err = decryptAESGCM(symKey, item.Value)
			if err != nil {
				logger.Fatal().Err(err).Msgf("error decrypting shared environment variable %s", item.Name)
			}
		}
		return envVar
	})

	envMap := lo.Associate(envVars, func(item DecryptedSecret) (string, string) {
		return item.Name, item.Value
	})

	personal := lo.Filter(envVars, func(item DecryptedSecret, index int) bool {
		return item.Personal
	})

	// join personal keys
	if len(personal) > 0 {
		secretsFile, err = readSecretsFile(env, true)
		if err != nil {
			logger.Fatal().Err(err).Msg("error reading personal secrets file, is it corrupted?")
		}

		personalEnvVars := lo.Map(secretsFile.Secrets, func(item EncryptedSecret, index int) DecryptedSecret {
			envVar := DecryptedSecret{
				Name:     item.Name,
				Value:    "",
				Personal: false,
			}
			envVar.Value, err = decryptAESGCM(symKey, item.Value)
			if err != nil {
				logger.Fatal().Err(err).Msgf("error decrypting personal environment variable %s", item.Name)
			}
			return envVar
		})

		for _, personalVar := range personalEnvVars {
			envMap[personalVar.Name] = personalVar.Value
		}
	}

	// Capture the existing env so we can
	existingEnv := os.Environ()
	existingEnvMap := lo.Associate(existingEnv, func(item string) (string, string) {
		parts := strings.SplitN(item, "=", 2)
		return parts[0], parts[1]
	})

	// generate source file
	var undoLines []string // so we can deactivate the env
	sourceFile := ""
	for key, val := range envMap {
		sourceFile += fmt.Sprintf("export %s=\"%s\"\n", key, val)
		// Check if the env already exists and make a replacement set
		if oldVal, exists := existingEnvMap[key]; exists {
			undoLines = append(undoLines, fmt.Sprintf("export %s=\"%s\"\n", key, oldVal))
			continue
		}

		// Otherwise unset the env var
		undoLines = append(undoLines, fmt.Sprintf("unset %s", key))
	}

	// We allow for nested prefixes so they user knows how they environments are stacked
	sourceFile += fmt.Sprintf("PS1=(epicenv: %s)$PS1\n", env)
	sourceFile += "deactivate() {\n  "
	sourceFile += strings.Join(undoLines, "\n  ")
	sourceFile += "}"

	tempSourcePath := path.Join(".epicenv", env, fmt.Sprintf("temp-%d", time.Now().UnixMilli()))
	err = os.WriteFile(tempSourcePath, []byte(sourceFile), 0777)
	if err != nil {
		logger.Fatal().Err(err).Msg("error writing out temp file")
	}

	fmt.Println(tempSourcePath)
}
