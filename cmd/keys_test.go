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

	keyPairs := findPrivateKeysForPublicKeys(pubKeys)
	if len(keyPairs) == 0 {
		t.Fatal("did not find any keys")
	}

	t.Log("found private key:", keyPairs)
}
