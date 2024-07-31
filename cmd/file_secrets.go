package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/samber/lo"
	"os"
	"path"
)

type (
	SecretsFile struct {
		Secrets []EncryptedSecret
	}
	EncryptedSecret struct {
		DecryptedSecret
		// Value is base64 encoded encrypted bytes
		Value string `json:",omitempty"`
	}
	DecryptedSecret struct {
		Name string
		// Value is the decrypted value
		Value string
		// Personal is whether this should be pulled from the personal_secrets.json file
		Personal bool
	}
)

func readSecretsFile(env string, personal bool) (*SecretsFile, error) {
	fileBytes, err := os.ReadFile(path.Join(".epicenv", env, lo.Ternary(personal, "personal_secrets.json", "secrets.json")))
	if err != nil {
		return nil, fmt.Errorf("error in os.ReadFile: %w", err)
	}

	var secretsFile SecretsFile
	err = json.Unmarshal(fileBytes, &secretsFile)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling keys file, is it corrupted?: %w", err)
	}

	return &secretsFile, nil
}

func writeSecretsFile(env string, keysFile SecretsFile, personal bool) error {
	fileBytes, err := json.Marshal(keysFile)
	if err != nil {
		return fmt.Errorf("error in json.Marshal: %w", err)
	}

	err = os.MkdirAll(path.Join(".epicenv", env), 0777)
	if err != nil {
		return fmt.Errorf("error in os.MkdirAll: %w", err)
	}

	err = os.WriteFile(path.Join(".epicenv", env, lo.Ternary(personal, "personal_secrets.json", "secrets.json")), fileBytes, 0777)
	if err != nil {
		return fmt.Errorf("error in os.WriteFile: %w", err)
	}

	return nil
}
