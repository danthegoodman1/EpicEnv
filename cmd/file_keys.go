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

		// IsHeadless indicates if this is a headless key (not associated with a GitHub user)
		IsHeadless bool
	}
)

func readKeysFile(env string) (*KeysFile, error) {
	fileBytes, err := os.ReadFile(path.Join(".epicenv", env, "keys.json"))
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
	fileBytes, err := json.MarshalIndent(keysFile, "", "  ")
	if err != nil {
		return fmt.Errorf("error in json.MarshalIndent: %w", err)
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
