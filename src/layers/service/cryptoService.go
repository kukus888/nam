package services

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"kukus/nam/v2/layers/data"

	"golang.org/x/crypto/pbkdf2"
)

// CryptoService handles encryption and decryption of secret data
type CryptoService struct {
	masterKey []byte
}

// NewCryptoService creates a new crypto service with a master key
func NewCryptoService(masterPassword string, salt []byte) *CryptoService {
	// Use PBKDF2 to derive a key from the master password
	key := pbkdf2.Key([]byte(masterPassword), salt, 100000, 32, sha256.New)
	return &CryptoService{
		masterKey: key,
	}
}

// Encrypt encrypts data using AES-256-GCM
func (c *CryptoService) Encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.masterKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// Decrypt decrypts data using AES-256-GCM
func (c *CryptoService) Decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.masterKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// EncryptSecretData encrypts secret data and returns encrypted bytes
func (c *CryptoService) EncryptSecretData(secretData data.SecretData) ([]byte, error) {
	plaintext, err := secretData.ToBytes()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize secret data: %w", err)
	}

	encrypted, err := c.Encrypt(plaintext)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt secret data: %w", err)
	}

	return encrypted, nil
}

// DecryptSecretData decrypts bytes and returns the appropriate SecretData type
func (c *CryptoService) DecryptSecretData(encryptedData []byte, secretType data.SecretType) (data.SecretData, error) {
	plaintext, err := c.Decrypt(encryptedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt secret data: %w", err)
	}

	// Create the appropriate secret data type
	secretData := data.SecretTypeFactory(secretType)

	// Unmarshal into the specific type
	if err := json.Unmarshal(plaintext, secretData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal secret data: %w", err)
	}

	return secretData, nil
}
