package cmd

import (
	"encoding/json"
	"errors"
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
		Name string
		// Value is base64 encoded encrypted bytes
		Value string `json:",omitempty"`
		// Personal is whether this should be pulled from the personal_secrets.json file
		Personal    bool
		Integration Integration `json:",omitempty"`
	}
	DecryptedSecret struct {
		Name string
		// Value is the decrypted value
		Value string
		// Personal is whether this should be pulled from the personal_secrets.json file
		Personal    bool
		Integration Integration `json:",omitempty"`
	}

	// Integration adds a level of indirection without changing the core functionality.
	//
	// It serves as an extra step to fetch the required secret.
	Integration string
)

const (
	// The secret is at the 1password referenced by the encrypted content
	Integration1Password Integration = "1password"
)

var knownIntegrations = []Integration{Integration1Password}

func readSecretsFile(env string, personal bool) (*SecretsFile, error) {
	fileBytes, err := os.ReadFile(path.Join(".epicenv", env, lo.Ternary(personal, "personal_secrets.json", "secrets.json")))
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
	// Verify we are in the right directory
	if !envExists(env) {
		logger.Fatal().Msgf("Environment %s not found, make sure there is a .epicenv directory in this directory", env)
	}

	fileBytes, err := json.Marshal(secretsFile)
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
