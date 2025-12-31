package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
)

type OverlayConfig struct {
	Base string `json:"base"`
}

func readOverlayConfig(env string) (*OverlayConfig, error) {
	epicEnvPath := getEpicEnvPath()
	fileBytes, err := os.ReadFile(path.Join(epicEnvPath, env, "overlay.json"))
	if err != nil {
		return nil, err
	}

	var config OverlayConfig
	err = json.Unmarshal(fileBytes, &config)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling overlay config: %w", err)
	}

	return &config, nil
}

func writeOverlayConfig(env string, config OverlayConfig) error {
	epicEnvPath := getEpicEnvPath()
	fileBytes, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("error in json.MarshalIndent: %w", err)
	}

	err = os.MkdirAll(path.Join(epicEnvPath, env), 0777)
	if err != nil {
		return fmt.Errorf("error in os.MkdirAll: %w", err)
	}

	err = os.WriteFile(path.Join(epicEnvPath, env, "overlay.json"), fileBytes, 0777)
	if err != nil {
		return fmt.Errorf("error in os.WriteFile: %w", err)
	}

	return nil
}

func isOverlay(env string) bool {
	_, err := readOverlayConfig(env)
	return err == nil
}

// resolveRootEnv recursively finds the ROOT non-overlay environment (for keys.json and invites)
func resolveRootEnv(env string) (string, error) {
	config, err := readOverlayConfig(env)
	if errors.Is(err, os.ErrNotExist) {
		return env, nil
	}
	if err != nil {
		return "", fmt.Errorf("error reading overlay config for %s: %w", env, err)
	}

	return resolveRootEnv(config.Base)
}

// getOverlayChain returns ordered slice from root to target, e.g. ["local", "testing", "agent-testing"]
func getOverlayChain(env string) ([]string, error) {
	config, err := readOverlayConfig(env)
	if errors.Is(err, os.ErrNotExist) {
		return []string{env}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error reading overlay config for %s: %w", env, err)
	}

	baseChain, err := getOverlayChain(config.Base)
	if err != nil {
		return nil, err
	}

	return append(baseChain, env), nil
}

