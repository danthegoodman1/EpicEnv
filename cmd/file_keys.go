package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
)

type (
	KeysFile struct {
		EncryptedKeys []EncryptedKey
	}

	EncryptedKey struct {
		// GitHub username, there may be many for the same
		Username  string
		PublicKey string

		// EncryptedSharedKey base64 encoded encrypted bytes
		EncryptedSharedKey string
	}
)

func readKeysFile(env string) (*KeysFile, error) {
	keysPath := path.Join(".epicenv", env, "keys.json")
	logger.Debug().Msgf("reading keys from %s", keysPath)
	fileBytes, err := os.ReadFile(keysPath)
	if err != nil {
		return nil, fmt.Errorf("error in os.ReadFile: %w", err)
	}

	var keysFile KeysFile
	err = json.Unmarshal(fileBytes, &keysFile)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling keys file, is it corrupted?: %w", err)
	}

	return &keysFile, nil
}

func writeKeysFile(env string, keysFile KeysFile) error {
	// Verify we are in the right directory
	if !envExists(env) {
		logger.Fatal().Msgf("Environment %s not found, make sure there is a .epicenv directory in this directory", env)
	}

	fileBytes, err := json.Marshal(keysFile)
	if err != nil {
		return fmt.Errorf("error in json.Marshal: %w", err)
	}

	err = os.MkdirAll(path.Join(".epicenv", env), 0777)
	if err != nil {
		return fmt.Errorf("error in os.MkdirAll: %w", err)
	}

	err = os.WriteFile(path.Join(".epicenv", env, "keys.json"), fileBytes, 0777)
	if err != nil {
		return fmt.Errorf("error in os.WriteFile: %w", err)
	}

	return nil
}
