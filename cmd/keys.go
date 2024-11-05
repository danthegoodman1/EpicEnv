package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/samber/lo"
	"golang.org/x/crypto/ssh"
)

var (
	ErrNotFound = errors.New("not found")
)

type keyPair struct {
	publicKeyContent string
	privateKeyPath   string
}

func findPrivateKeysForPublicKeys(pubKeys []string) []keyPair {
	files, err := os.ReadDir(path.Join(os.Getenv("HOME"), ".ssh"))
	if err != nil {
		return nil
	}

	var keys []keyPair

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// Find a file that shares content with the pub keys
		fileContent, err := os.ReadFile(path.Join(os.Getenv("HOME"), ".ssh", file.Name()))
		if err != nil {
			return nil
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
			return nil
		}

		keys = append(keys, keyPair{
			publicKeyContent: keyContent,
			privateKeyPath:   path.Join(os.Getenv("HOME"), ".ssh", privateKeyFile),
		})
	}

	return keys
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

	if res.StatusCode == 404 {
		return nil, ErrNotFound
	}

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error in io.ReadAll(res.Body): %w", err)
	}

	if res.StatusCode >= 299 {
		return nil, fmt.Errorf("high status code: %d %s", err, string(bodyBytes))
	}

	keys := lo.Map(strings.Split(string(bodyBytes), "\n"), func(item string, index int) string {
		return strings.TrimSpace(item)
	})
	keys = lo.Filter(keys, func(item string, index int) bool {
		return item != "" && strings.HasPrefix(item, "ssh-rsa")
	})

	return keys, nil
}

// Add this function to handle machine user key loading
func findPrivateKeyForMachineName(machineName string, pubKeyContent string) (keyPair, error) {
	// First check if EPICENV_MACHINE_KEY is set
	if keyPath := os.Getenv("EPICENV_MACHINE_KEY"); keyPath != "" {
		fileContent, err := os.ReadFile(keyPath)
		if err != nil {
			return keyPair{}, fmt.Errorf("error reading machine key file: %w", err)
		}

		// Verify this is the correct key by comparing public keys
		pubKey, err := ssh.ParsePrivateKey(fileContent)
		if err != nil {
			return keyPair{}, fmt.Errorf("error parsing private key: %w", err)
		}

		marshaledPubKey := string(ssh.MarshalAuthorizedKey(pubKey.PublicKey()))
		if marshaledPubKey == pubKeyContent {
			return keyPair{
				publicKeyContent: pubKeyContent,
				privateKeyPath:   keyPath,
			}, nil
		}
		return keyPair{}, fmt.Errorf("provided key does not match machine user's public key")
	}

	// If no env var, look in default locations
	defaultLocations := []string{
		path.Join(os.Getenv("HOME"), ".ssh", fmt.Sprintf("machine_%s", machineName)),
		path.Join(os.Getenv("HOME"), ".ssh", machineName),
	}

	for _, loc := range defaultLocations {
		if _, err := os.Stat(loc); err == nil {
			// Found a potential key, verify it
			fileContent, err := os.ReadFile(loc)
			if err != nil {
				continue
			}

			pubKey, err := ssh.ParsePrivateKey(fileContent)
			if err != nil {
				continue
			}

			marshaledPubKey := string(ssh.MarshalAuthorizedKey(pubKey.PublicKey()))
			if marshaledPubKey == pubKeyContent {
				return keyPair{
					publicKeyContent: pubKeyContent,
					privateKeyPath:   loc,
				}, nil
			}
		}
	}

	return keyPair{}, fmt.Errorf("could not find matching private key for machine user %s", machineName)
}
