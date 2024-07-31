package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
)

type (
	SecretsFile struct {
		Secrets []EncryptedSecret
	}
	EncryptedSecret struct {
		Name string
		// Value is base64 encoded encrypted bytes
		Value string
		// Personal is whether this should be pulled from the personal_secrets.json file
		Personal bool
	}
)

func readSecretsFile(env string) (*KeysFile, error) {
	fileBytes, err := os.ReadFile(path.Join(".epicenv", env, "secrets.json"))
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

func writeSecretsFile(env string, keysFile KeysFile) error {
	fileBytes, err := json.Marshal(keysFile)
	if err != nil {
		return fmt.Errorf("error in json.Marshal: %w", err)
	}

	err = os.MkdirAll(path.Join(".epicenv", env), 0777)
	if err != nil {
		return fmt.Errorf("error in os.MkdirAll: %w", err)
	}

	err = os.WriteFile(path.Join(".epicenv", env, "secrets.json"), fileBytes, 0777)
	if err != nil {
		return fmt.Errorf("error in os.WriteFile: %w", err)
	}

	return nil
}
