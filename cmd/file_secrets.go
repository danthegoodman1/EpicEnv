package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/samber/lo"
)

type (
	SecretsFile struct {
		Secrets []EncryptedSecret
	}
	EncryptedSecret struct {
		Name string
		// Value is base64 encoded encrypted bytes
		Value string `json:",omitempty"`
		// Personal is whether this should be pulled from the personal_secrets.json file
		Personal bool
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
	epicEnvPath := getEpicEnvPath()
	fileBytes, err := os.ReadFile(path.Join(epicEnvPath, env, lo.Ternary(personal, "personal_secrets.json", "secrets.json")))
	if personal && errors.Is(err, os.ErrNotExist) {
		// Create a blank one and return
		secretsFile := SecretsFile{}
		return &secretsFile, writeSecretsFile(env, secretsFile, true)
	}

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

func writeSecretsFile(env string, secretsFile SecretsFile, personal bool) error {
	epicEnvPath := getEpicEnvPath()
	fileBytes, err := json.MarshalIndent(secretsFile, "", "  ")
	if err != nil {
		return fmt.Errorf("error in json.MarshalIndent: %w", err)
	}

	err = os.MkdirAll(path.Join(epicEnvPath, env), 0777)
	if err != nil {
		return fmt.Errorf("error in os.MkdirAll: %w", err)
	}

	err = os.WriteFile(path.Join(epicEnvPath, env, lo.Ternary(personal, "personal_secrets.json", "secrets.json")), fileBytes, 0777)
	if err != nil {
		return fmt.Errorf("error in os.WriteFile: %w", err)
	}

	return nil
}
