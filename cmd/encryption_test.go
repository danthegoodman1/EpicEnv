package cmd

import (
	"bytes"
	"os"
	"path"
	"testing"
)

// TestEncryptDecrypt tests the encryption and decryption process with RSA and ED25519 keys.
func TestEncryptionDecryption(t *testing.T) {
	// Replace these with the actual paths to your test keys
	rsaContent, err := os.ReadFile(path.Join(os.Getenv("HOME"), ".ssh", "id_rsa.pub"))
	if err != nil {
		t.Fatal(err)
	}
	rsaKeyPair := keyPair{
		publicKeyContent: string(rsaContent),                             // Replace with your actual public key content
		privateKeyPath:   path.Join(os.Getenv("HOME"), ".ssh", "id_rsa"), // Replace with the actual path to your private key
	}

	testData := []byte("Hello, World!")

	// Test RSA encryption and decryption
	encryptedRSA, err := encryptWithPublicKey(testData, rsaKeyPair.publicKeyContent)
	if err != nil {
		t.Fatalf("Failed to encrypt with RSA key: %v", err)
	}

	decryptedRSA, err := decryptWithPrivateKey(encryptedRSA, rsaKeyPair)
	if err != nil {
		t.Fatalf("Failed to decrypt with RSA key: %v", err)
	}

	if !bytes.Equal(testData, decryptedRSA) {
		t.Errorf("RSA decrypted data does not match original data")
	}
}
