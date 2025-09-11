package services

import (
	"fmt"
	"log/slog"

	"kukus/nam/v2/layers/data"

	"github.com/jackc/pgx/v5/pgxpool"
)

// SecretsService provides high-level operations for managing secrets
type SecretsService struct {
	db     *pgxpool.Pool
	logger *slog.Logger
	crypto *CryptoService
}

// NewSecretsService creates a new secrets service
func NewSecretsService(db *pgxpool.Pool, logger *slog.Logger, crypto *CryptoService) *SecretsService {
	return &SecretsService{
		db:     db,
		logger: logger,
		crypto: crypto,
	}
}

// CreateSecret creates and stores a new encrypted secret
func (s *SecretsService) CreateSecret(dto *data.SecretDTO, userId *uint64) (*uint64, error) {
	// Encrypt and set the secret data
	encryptedData, err := s.crypto.Encrypt([]byte(dto.Data))
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt secret: %w", err)
	}
	secret, err := dto.ToSecret(encryptedData)
	if err != nil {
		return nil, fmt.Errorf("failed to create secret from DTO: %w", err)
	}

	// Set audit fields
	secret.CreatedBy = userId
	secret.UpdatedBy = userId

	// Insert into database
	dao := dto.ToSecretDAO(encryptedData)
	id, err := dao.DbInsert(s.db)
	if err != nil {
		return nil, fmt.Errorf("failed to store secret: %w", err)
	}

	s.logger.Info("Secret created", "id", *id, "name", dto.Name, "type", dto.Type)
	return id, nil
}

// GetSecret retrieves and decrypts a secret by ID
func (s *SecretsService) GetSecret(id uint64) (*data.Secret, error) {
	panic("Not implemented yet")
	/*
		secret, err := data.GetSecretById(s.db, id)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to retrieve secret: %w", err)
		}


		// Decrypt the secret data

			secretData, err := secret.GetSecretData()
			if err != nil {
				return nil, nil, fmt.Errorf("failed to decrypt secret: %w", err)
			}

			return secret, secretData, nil
	*/
}

// GetSecretsMetadata returns all secrets metadata (no decrypted data)
func (s *SecretsService) GetSecretsMetadata() ([]data.SecretDAO, error) {
	secrets, err := data.GetAllSecrets(s.db)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve secrets: %w", err)
	}

	// Clear the encrypted data for security (only return metadata)
	for i := range secrets {
		secrets[i].Data = nil
	}

	return secrets, nil
}

// DeleteSecret deletes a secret by ID
func (s *SecretsService) DeleteSecret(id uint64) error {
	if err := data.DeleteSecret(s.db, id); err != nil {
		return fmt.Errorf("failed to delete secret: %w", err)
	}

	s.logger.Info("Secret deleted", "id", id)
	return nil
}

// UpdateSecret updates an existing secret
func (s *SecretsService) UpdateSecret(id uint64, dto *data.SecretDTO, userId *uint64) error {
	panic("Not implemented yet")
	/*
		// Validate the secret data
		if err := dto.Data.Validate(); err != nil {
			return fmt.Errorf("invalid secret data: %w", err)
		}

		// Get existing secret
		secret, err := data.GetSecretById(s.db, id)
		if err != nil {
			return fmt.Errorf("secret not found: %w", err)
		}

		// Update fields
		secret.Name = dto.Name
		secret.Description = dto.Description
		secret.Type = string(dto.Type)
		secret.Metadata = dto.Metadata
		secret.UpdatedBy = userId

		// Encrypt and set the new secret data
		if err := secret.SetSecretData(dto.Data); err != nil {
			return fmt.Errorf("failed to encrypt secret: %w", err)
		}

		// Update in database
		if err := data.UpdateSecret(s.db, id, secret); err != nil {
			return fmt.Errorf("failed to update secret: %w", err)
		}

		s.logger.Info("Secret updated", "id", id, "name", dto.Name, "type", dto.Type)
		return nil
	*/
}
