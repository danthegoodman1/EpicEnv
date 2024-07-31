package cmd

import (
	"testing"
)

const ghUsername = "danthegoodman1"

func TestGetKeysForGitHubUsername(t *testing.T) {
	keys, err := getKeysForGithubUsername(ghUsername)
	if err != nil {
		t.Fatal(err)
	}
	if len(keys) < 1 {
		t.Fatalf("Got no keys for %s!", ghUsername)
	}
	t.Log(keys)
}

func TestFindFirstPrivateKeyForPublicKey(t *testing.T) {
	pubKeys, err := getKeysForGithubUsername(ghUsername)
	if err != nil {
		t.Fatal(err)
	}

	privKeyName, err := findFirstPrivateKeyForPublicKeys(pubKeys)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("found private key:", privKeyName)
}
