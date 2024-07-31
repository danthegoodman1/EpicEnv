package cmd

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"io/ioutil"
)

func generateAESKey() []byte {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		logger.Fatal().Err(err).Msg("error generating random bytes")
	}
	return key
}

// Encrypt a string using AES-GCM and return the base64-encoded result
func encryptAESGCM(key []byte, plaintext string) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt a base64-encoded string using AES-GCM
func decryptAESGCM(key []byte, ciphertextBase64 string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextBase64)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// Returns a base64 encoded string
func encryptWithPublicKey(data []byte, publicKey string) (string, error) {
	// Parse the public key
	pub, _, _, _, err := ssh.ParseAuthorizedKey([]byte(publicKey))
	if err != nil {
		return "", err
	}

	// Convert the SSH public key to a crypto/rsa PublicKey
	cryptoPub, ok := pub.(ssh.CryptoPublicKey)
	if !ok {
		return "", errors.New("SSH key is not an RSA key")
	}
	rsaPub, ok := cryptoPub.CryptoPublicKey().(*rsa.PublicKey)
	if !ok {
		return "", errors.New("SSH key is not an RSA key")
	}

	// Encrypt the data
	encryptedData, err := rsa.EncryptPKCS1v15(rand.Reader, rsaPub, data)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(encryptedData), nil
}

// decryptWithPrivateKey decrypts data using the private key in the keyPair.
func decryptWithPrivateKey(encryptedData string, kp keyPair) ([]byte, error) {
	privateKeyBytes, err := ioutil.ReadFile(kp.privateKeyPath)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(privateKeyBytes)
	if block == nil {
		return nil, errors.New("failed to decode PEM block containing private key")
	}

	var privateKey interface{}
	if block.Type == "RSA PRIVATE KEY" {
		privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
	} else if block.Type == "OPENSSH PRIVATE KEY" {
		privateKey, err = ssh.ParseRawPrivateKey(privateKeyBytes)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("unsupported private key type")
	}

	decodedData, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return nil, err
	}

	switch key := privateKey.(type) {
	case *rsa.PrivateKey:
		return rsa.DecryptPKCS1v15(rand.Reader, key, decodedData)
	case *ed25519.PrivateKey:
		return nil, errors.New("decryption with Ed25519 is not supported")
	default:
		return nil, errors.New("unsupported private key type")
	}
}
