package services

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"kukus/nam/v2/layers/data"
	"sync"

	"golang.org/x/crypto/pbkdf2"
)

// CryptoService handles encryption and decryption of secret data
type CryptoService struct {
	masterKey []byte
}

var cryptoLock = &sync.Once{}
var cryptoService *CryptoService

// GetCryptoService returns the singleton instance of CryptoService.
// It panics if the service has not been initialized.
func GetCryptoService() *CryptoService {
	if cryptoService == nil {
		cryptoLock.Do(func() {
			panic("CryptoService not initialized. Call NewCryptoService first.")
		})
	}
	return cryptoService
}

// NewCryptoService creates a new crypto service with a master key
// Used for encrypting and decrypting secrets into the database
func NewCryptoService(masterPassword string, salt []byte) {
	cryptoLock.Do(func() {
		// Use PBKDF2 to derive a key from the master password
		key := pbkdf2.Key([]byte(masterPassword), salt, 100000, 32, sha256.New)
		cryptoService = &CryptoService{
			masterKey: key,
		}
	})
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

// EncryptSecret encrypts the data in a Secret and returns a SecretDAO with encrypted data
func (c *CryptoService) EncryptSecret(secret *data.Secret) (*data.SecretDAO, error) {
	encryptedData, err := c.Encrypt(secret.Data)
	if err != nil {
		return nil, err
	}
	secretDAO := secret.ToSecretDAO(encryptedData)
	return secretDAO, nil
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

// DecryptDAO decrypts the data in a SecretDAO and returns a Secret with plaintext data
func (c *CryptoService) DecryptDAO(dao *data.SecretDAO) (*data.Secret, error) {
	plaintext, err := c.Decrypt(dao.Data)
	if err != nil {
		return nil, err
	}
	return &data.Secret{
		Id:          dao.Id,
		Type:        dao.Type,
		Name:        dao.Name,
		Description: dao.Description,
		Data:        plaintext,
		Metadata:    dao.Metadata,
		CreatedAt:   dao.CreatedAt,
		UpdatedAt:   dao.UpdatedAt,
		CreatedBy:   dao.CreatedBy,
		UpdatedBy:   dao.UpdatedBy,
	}, nil
}
