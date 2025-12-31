package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

// getEnvOrFlag will attempt to read the flag, then environment, then auto-infer if only one exists
func getEnvOrFlag(cmd *cobra.Command) string {
	if env := cmd.Flag("environment").Value.String(); env != "" {
		return env
	}

	if env := os.Getenv("EPICENV"); env != "" {
		return env
	}

	// Try to auto-infer if there's only one environment
	environments, err := listEnvironments()
	if err != nil {
		logger.Fatal().Err(err).Msg("Error listing environments")
	}

	if len(environments) == 1 {
		logger.Debug().Msgf("Auto-inferred environment: %s", environments[0])
		return environments[0]
	}

	if len(environments) == 0 {
		logger.Fatal().Msg("No environments found. Run 'epicenv init' to create one")
	} else {
		logger.Fatal().Msgf("Multiple environments found (%s), please specify with -e", strings.Join(environments, ", "))
	}

	os.Exit(1)
	return ""
}

type loadedEnvVar struct {
	Value    string
	Personal bool
}

// loadEnv will short circuit fatal exit if it has an unrecoverable error.
// For overlay environments, it loads and merges secrets through the entire chain.
func loadEnv(env string) map[string]loadedEnvVar {
	symKey, err := loadSymmetricKey(env)
	if err != nil {
		logger.Fatal().Err(err).Msg("error loading symmetric key")
	}

	// Get the overlay chain (from root to target)
	chain, err := getOverlayChain(env)
	if err != nil {
		logger.Fatal().Err(err).Msg("error getting overlay chain")
	}

	envMap := make(map[string]loadedEnvVar)

	// Load and merge secrets from each environment in the chain
	for _, chainEnv := range chain {
		loadEnvLayer(chainEnv, symKey, envMap)
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

// loadEnvLayer loads secrets from a single environment and merges them into envMap.
// Later layers override earlier ones.
func loadEnvLayer(env string, symKey []byte, envMap map[string]loadedEnvVar) {
	secretsFile, err := readSecretsFile(env, false)
	if errors.Is(err, os.ErrNotExist) {
		return
	}
	if err != nil {
		logger.Fatal().Err(err).Msgf("error reading shared secrets file for %s, is it corrupted?", env)
	}

	// Track which keys are personal in this layer (need personal values)
	var personalKeys []string

	for _, item := range secretsFile.Secrets {
		if item.Personal {
			personalKeys = append(personalKeys, item.Name)
			// Mark as personal placeholder if not already set with a value
			if existing, exists := envMap[item.Name]; !exists || existing.Value == "" {
				envMap[item.Name] = loadedEnvVar{
					Value:    "",
					Personal: true,
				}
			}
		} else {
			decrypted, err := decryptAESGCM(symKey, item.Value)
			if err != nil {
				logger.Fatal().Err(err).Msgf("error decrypting shared environment variable %s", item.Name)
			}
			envMap[item.Name] = loadedEnvVar{
				Value:    decrypted,
				Personal: false,
			}
		}
	}

	// Load personal secrets for this layer
	if len(personalKeys) > 0 {
		personalSecretsFile, err := readSecretsFile(env, true)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			logger.Fatal().Err(err).Msgf("error reading personal secrets file for %s, is it corrupted?", env)
		}

		if personalSecretsFile != nil {
			for _, item := range personalSecretsFile.Secrets {
				decrypted, err := decryptAESGCM(symKey, item.Value)
				if err != nil {
					logger.Fatal().Err(err).Msgf("error decrypting personal environment variable %s", item.Name)
				}
				envMap[item.Name] = loadedEnvVar{
					Value:    decrypted,
					Personal: true,
				}
			}
		}
	}
}

func loadSymmetricKey(env string) ([]byte, error) {
	// Resolve to root environment for overlays
	rootEnv, err := resolveRootEnv(env)
	if err != nil {
		return nil, fmt.Errorf("error resolving root environment: %w", err)
	}

	keysFile, err := readKeysFile(rootEnv)
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
		return nil, fmt.Errorf("did not find the known public key again among the encrypted keys, this is a bug. Please report")
	}

	symKey, err := decryptWithPrivateKey(actualKey.EncryptedSharedKey, chosenKey)
	if err != nil {
		return nil, err
	}

	return symKey, nil
}
