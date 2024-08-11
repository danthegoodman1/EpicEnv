package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/1password/onepassword-sdk-go"
	"os"
	"path"
	"time"
)

type OnePasswordIntegrationFile struct {
	EncryptedServiceAccount string
}

func Setup1PasswordServiceAccount(env string, serviceAccountToken string) error {
	symKey, err := loadSymmetricKey(env)
	if err != nil {
		return fmt.Errorf("error loading symmetric key: %w", err)
	}

	// Store the encrypted value
	encryptedServiceAccount, err := encryptAESGCM(symKey, serviceAccountToken)
	if err != nil {
		return fmt.Errorf("error encrypting service account token: %w", err)
	}

	fileContent, err := json.Marshal(OnePasswordIntegrationFile{
		EncryptedServiceAccount: encryptedServiceAccount,
	})
	if err != nil {
		return fmt.Errorf("error in json.Marshal: %w", err)
	}

	err = os.WriteFile(path.Join(".epicenv", env, "1password.json"), fileContent, 0777)
	if err != nil {
		return fmt.Errorf("error writing 1password.json: %w", err)
	}

	return nil
}

func Read1PasswordServiceAccount(env string) (string, error) {
	symKey, err := loadSymmetricKey(env)
	if err != nil {
		return "", fmt.Errorf("error loading symmetric key: %w", err)
	}

	fileContent, err := os.ReadFile(path.Join(".epicenv", env, "1password.json"))
	if err != nil {
		return "", fmt.Errorf("error reading 1password.json: %w", err)
	}

	var integrationFile OnePasswordIntegrationFile
	err = json.Unmarshal(fileContent, &integrationFile)
	if err != nil {
		return "", fmt.Errorf("error unmarshaling json integration file: %w", err)
	}

	decrypted, err := decryptAESGCM(symKey, integrationFile.EncryptedServiceAccount)
	if err != nil {
		return "", fmt.Errorf("error decrypting service account: %w", err)
	}

	return decrypted, nil
}

func Load1PasswordSecret(env, secretReference string) string {
	logger.Debug().Msgf("reading secret reference %s from 1Password", secretReference)
	serviceAccount, err := Read1PasswordServiceAccount(env)
	if err != nil {
		err = fmt.Errorf("error in Read1PasswordServiceAccount: %w", err)
		logger.Error().Err(err).Msg("error loading 1Password secret, failing through (env var will be error msg)")
		return "ERROR see terminal output"
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	client, err := onepassword.NewClient(
		ctx,
		onepassword.WithServiceAccountToken(serviceAccount),
		onepassword.WithIntegrationInfo("EpicEnv", Version),
	)
	if err != nil {
		err = fmt.Errorf("error in onepassword.NewClient: %w", err)
		logger.Error().Err(err).Msg("error loading 1Password secret, failing through (env var will be error msg)")
		return err.Error()
	}

	secret, err := client.Secrets.Resolve(ctx, secretReference)
	if err != nil {
		err = fmt.Errorf("error in client.Secrets.Resolve: %w", err)
		logger.Error().Err(err).Msg("error loading 1Password secret, failing through (env var will be error msg)")
		return err.Error()
	}

	return secret
}
