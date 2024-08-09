package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func Test1PasswordSetupAndLoad(t *testing.T) {
	// Just verify that we can set and get it back
	token := "testoken"
	env := "default"
	{
		ex, err := os.Executable()
		if err != nil {
			panic(err)
		}
		exPath := filepath.Dir(ex)
		fmt.Println(exPath)
	}
	err := Setup1PasswordServiceAccount(env, token)
	if err != nil {
		t.Fatal(err)
	}

	readToken, err := Read1PasswordServiceAccount(env)
	if err != nil {
		t.Fatal(err)
	}

	if readToken != token {
		t.Fatalf("Did not get matching tokens, got: %s", readToken)
	}
}
