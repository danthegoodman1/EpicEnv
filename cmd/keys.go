package cmd

import (
	"context"
	"errors"
	"fmt"
	"github.com/samber/lo"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

var (
	ErrNotFound = errors.New("not found")
)

func findFirstPrivateKeyForPublicKeys(pubKeys []string) (string, error) {
	files, err := os.ReadDir(path.Join(os.Getenv("HOME"), ".ssh"))
	if err != nil {
		return "", fmt.Errorf("error in ReadDir: %w", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// Find a file that shares content with the pub keys
		fileContent, err := os.ReadFile(path.Join(os.Getenv("HOME"), ".ssh", file.Name()))
		if err != nil {
			return "", fmt.Errorf("error in os.ReadFile of %s: %w", file.Name(), err)
		}

		logger.Debug().Msgf("Checking file %s", file.Name())

		keyParts := strings.Split(string(fileContent), " ")
		if len(keyParts) < 2 {
			continue
		}
		keyContent := strings.Join(keyParts[:2], " ") // remove any hostname info after

		if !lo.Contains(pubKeys, keyContent) {
			continue
		}

		// We found a match
		// pop the .pub
		privateKeyFile := strings.Split(file.Name(), ".pub")[0]
		if privateKeyFile == file.Name() {
			return "", fmt.Errorf("private key file was same as public key, or rather {key_file}.pub did not have a matching {key_file} file")
		}

		return path.Join(os.Getenv("HOME"), ".ssh", privateKeyFile), nil
	}

	return "", ErrNotFound
}

func getKeysForGithubUsername(username string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://github.com/%s.keys", username), nil)
	if err != nil {
		return nil, fmt.Errorf("error in http.NewRequestWithContext: %w", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error in http.DefaultClient.Do: %w", err)
	}
	defer res.Body.Close()

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error in io.ReadAll(res.Body): %w", err)
	}

	keys := lo.Map(strings.Split(string(bodyBytes), "\n"), func(item string, index int) string {
		return strings.TrimSpace(item)
	})

	return keys, nil
}
