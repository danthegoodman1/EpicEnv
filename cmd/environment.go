package cmd

import (
	"errors"
	"fmt"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

// getEnvOrFlag will attempt to read the flag, then environment, then print a message and exit
func getEnvOrFlag(cmd *cobra.Command) string {
	if env := cmd.Flag("environment").Value.String(); env != "" {
		return env
	}

	if env := os.Getenv("EPICENV"); env != "" {
		return env
	}

	logger.Fatal().Msg("Could not infer environment, please specify with -e")
	os.Exit(1)
	return ""
}

type loadedEnvVar struct {
	Value    string
	Personal bool
}

// loadEnv will short circuit fatal exit if it has an unrecoverable error
func loadEnv(env string) map[string]loadedEnvVar {
	symKey, err := loadSymmetricKey(env)
	if err != nil {
		logger.Fatal().Err(err).Msg("error reading key file")
	}

	// decrypt keys
	secretsFile, err := readSecretsFile(env, false)
	if errors.Is(err, os.ErrNotExist) {
		return map[string]loadedEnvVar{}
	}
	if err != nil {
		logger.Fatal().Err(err).Msg("error reading shared secrets file, is it corrupted?")
	}

	envVars := lo.Map(secretsFile.Secrets, func(item EncryptedSecret, index int) DecryptedSecret {
		envVar := DecryptedSecret{
			Name:     item.Name,
			Value:    "",
			Personal: item.Personal,
		}
		if !item.Personal {
			envVar.Value, err = decryptAESGCM(symKey, item.Value)
			if err != nil {
				logger.Fatal().Err(err).Msgf("error decrypting shared environment variable %s", item.Name)
			}
		}
		return envVar
	})

	envMap := lo.Associate(envVars, func(item DecryptedSecret) (string, loadedEnvVar) {
		return item.Name, loadedEnvVar{
			Value:    item.Value,
			Personal: item.Personal,
		}
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
			envMap[personalVar.Name] = loadedEnvVar{
				Value:    personalVar.Value,
				Personal: true,
			}
		}
	}

	// Find any that we didn't fill in from personal secrets and warn
	missingPersonal := lo.PickBy(envMap, func(key string, value loadedEnvVar) bool {
		return value.Personal && value.Value == ""
	})
	if len(missingPersonal) > 0 {
		logger.Warn().Msgf("Missing personal values: %s", strings.Join(lo.Keys(missingPersonal), ", "))
	}

	return envMap
}

func loadSymmetricKey(env string) ([]byte, error) {
	keysFile, err := readKeysFile(env)
	if errors.Is(err, os.ErrNotExist) {
		logger.Fatal().Msg("Keys file not found, make sure to run init command first")
	}
	if err != nil {
		return nil, fmt.Errorf("error reading keys file: %w", err)
	}

	keyPairs := findPrivateKeysForPublicKeys(lo.Map(keysFile.EncryptedKeys, func(item EncryptedKey, index int) string {
		return item.PublicKey
	}))

	if len(keyPairs) == 0 {
		return nil, fmt.Errorf("did not find any local private keys matching a known public key, are you invited to the %s environment?", env)
	}

	// Use the first one we found
	chosenKey := keyPairs[0]

	// decrypt symmetric key
	actualKey, found := lo.Find(keysFile.EncryptedKeys, func(item EncryptedKey) bool {
		return item.PublicKey == chosenKey.publicKeyContent
	})
	if !found {
		return nil, fmt.Errorf("did not find the known public key again among the encrypted keys, this is a bug. Please report.")
	}

	symKey, err := decryptWithPrivateKey(actualKey.EncryptedSharedKey, chosenKey)
	if err != nil {
		return nil, err
	}

	return symKey, nil
}
