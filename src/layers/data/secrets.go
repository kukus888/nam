package data

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Secret represents a secret business object
type SecretDAO struct {
	Id          uint64                 `json:"id" db:"id"`
	Type        string                 `json:"type" db:"type"`
	Name        string                 `json:"name" db:"name"`
	Description string                 `json:"description" db:"description"`
	Data        []byte                 `json:"-" db:"data"` // Encrypted data, excluded from JSON
	Metadata    map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
	CreatedBy   *uint64                `json:"created_by" db:"created_by"`
	UpdatedBy   *uint64                `json:"updated_by" db:"updated_by"`
}

// Secret represents a secret business object
type Secret struct {
	Id          uint64                 `json:"id"`
	Type        string                 `json:"type"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Data        []byte                 `json:"-"` // Decrypted data, Excluded from JSON
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	CreatedBy   *uint64                `json:"created_by"`
	UpdatedBy   *uint64                `json:"updated_by"`
}

// SecretDTO for creating/updating secrets
type SecretDTO struct {
	Id          uint64                 `json:"id"`
	Type        string                 `json:"type" binding:"required"`
	Name        string                 `json:"name" binding:"required"`
	Description string                 `json:"description"`
	Data        string                 `json:"data" binding:"required"` // Decrypted data
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	CreatedBy   *uint64                `json:"created_by"`
	UpdatedBy   *uint64                `json:"updated_by"`
}

func (s *Secret) ToSecretDAO(encryptedData []byte) *SecretDAO {
	secret := &SecretDAO{
		Id:          s.Id,
		Type:        s.Type,
		Name:        s.Name,
		Description: s.Description,
		Metadata:    s.Metadata,
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
		CreatedBy:   s.CreatedBy,
		UpdatedBy:   s.UpdatedBy,
		Data:        encryptedData,
	}

	return secret
}

func (s *SecretDTO) ToSecret() (*Secret, error) {
	secret := &Secret{
		Id:          s.Id,
		Type:        s.Type,
		Name:        s.Name,
		Description: s.Description,
		Metadata:    s.Metadata,
		Data:        []byte(s.Data),
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
		CreatedBy:   s.CreatedBy,
		UpdatedBy:   s.UpdatedBy,
	}
	return secret, nil
}

func (s *Secret) ToDTO() *SecretDTO {
	return &SecretDTO{
		Id:          s.Id,
		Type:        s.Type,
		Name:        s.Name,
		Description: s.Description,
		Data:        string(s.Data),
		Metadata:    s.Metadata,
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
		CreatedBy:   s.CreatedBy,
		UpdatedBy:   s.UpdatedBy,
	}
}

// Database operations for secrets

func (s *SecretDAO) DbInsert(pool *pgxpool.Pool) (*uint64, error) {
	tx, err := pool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(context.Background())

	metadataJSON, err := json.Marshal(s.Metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	err = tx.QueryRow(context.Background(), `
		INSERT INTO secrets (type, name, description, data, metadata, created_by, updated_by) 
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id;
	`, s.Type, s.Name, s.Description, s.Data, metadataJSON, s.CreatedBy, s.UpdatedBy).Scan(&s.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to insert secret: %w", err)
	}

	return &s.Id, tx.Commit(context.Background())
}

func GetSecretById(pool *pgxpool.Pool, id uint64) (*SecretDAO, error) {
	tx, err := pool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(context.Background())

	var secret SecretDAO
	var metadataJSON []byte

	err = tx.QueryRow(context.Background(), `
		SELECT id, type, name, description, data, metadata, created_at, updated_at, created_by, updated_by 
		FROM secrets WHERE id = $1
	`, id).Scan(&secret.Id, &secret.Type, &secret.Name, &secret.Description,
		&secret.Data, &metadataJSON, &secret.CreatedAt, &secret.UpdatedAt,
		&secret.CreatedBy, &secret.UpdatedBy)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret: %w", err)
	}

	if err := json.Unmarshal(metadataJSON, &secret.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &secret, tx.Commit(context.Background())
}

func GetSecretByName(pool *pgxpool.Pool, name string) (*SecretDAO, error) {
	tx, err := pool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(context.Background())

	var secret SecretDAO
	var metadataJSON []byte

	err = tx.QueryRow(context.Background(), `
		SELECT id, type, name, description, data, metadata, created_at, updated_at, created_by, updated_by 
		FROM secrets WHERE name = $1
	`, name).Scan(&secret.Id, &secret.Type, &secret.Name, &secret.Description,
		&secret.Data, &metadataJSON, &secret.CreatedAt, &secret.UpdatedAt,
		&secret.CreatedBy, &secret.UpdatedBy)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret by name: %w", err)
	}

	if err := json.Unmarshal(metadataJSON, &secret.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &secret, tx.Commit(context.Background())
}

func GetAllSecrets(pool *pgxpool.Pool) ([]SecretDAO, error) {
	tx, err := pool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(context.Background())

	rows, err := tx.Query(context.Background(), `
		SELECT id, type, name, description, data, metadata, created_at, updated_at, created_by, updated_by 
		FROM secrets ORDER BY type, name
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query all secrets: %w", err)
	}
	defer rows.Close()

	var secrets []SecretDAO
	for rows.Next() {
		var secret SecretDAO
		var metadataJSON []byte

		err := rows.Scan(&secret.Id, &secret.Type, &secret.Name, &secret.Description,
			&secret.Data, &metadataJSON, &secret.CreatedAt, &secret.UpdatedAt,
			&secret.CreatedBy, &secret.UpdatedBy)
		if err != nil {
			return nil, fmt.Errorf("failed to scan secret row: %w", err)
		}

		if err := json.Unmarshal(metadataJSON, &secret.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		secrets = append(secrets, secret)
	}

	return secrets, tx.Commit(context.Background())
}

func UpdateSecret(pool *pgxpool.Pool, id uint64, secret *SecretDAO) error {
	tx, err := pool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	metadataJSON, err := json.Marshal(secret.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	_, err = tx.Exec(context.Background(), `
		UPDATE secrets 
		SET type = $1, name = $2, description = $3, data = $4, metadata = $5, updated_at = NOW(), updated_by = $6
		WHERE id = $7
	`, secret.Type, secret.Name, secret.Description, secret.Data, metadataJSON, secret.UpdatedBy, id)
	if err != nil {
		return fmt.Errorf("failed to update secret: %w", err)
	}

	return tx.Commit(context.Background())
}

func DeleteSecret(pool *pgxpool.Pool, id uint64) error {
	tx, err := pool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	_, err = tx.Exec(context.Background(), `DELETE FROM secrets WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete secret: %w", err)
	}

	return tx.Commit(context.Background())
}

// Get secrets that can be used for SSH authentication (passwords and private keys)
func GetSshSecrets(pool *pgxpool.Pool) ([]SecretDAO, error) {
	tx, err := pool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(context.Background())

	rows, err := tx.Query(context.Background(), `
		SELECT id, type, name, description, data, metadata, created_at, updated_at, created_by, updated_by 
		FROM secrets 
		WHERE type IN ('password', 'private_key', 'ssh_key')
		ORDER BY type, name
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query SSH secrets: %w", err)
	}
	defer rows.Close()

	var secrets []SecretDAO
	for rows.Next() {
		var secret SecretDAO
		var metadataJSON []byte

		err := rows.Scan(&secret.Id, &secret.Type, &secret.Name, &secret.Description,
			&secret.Data, &metadataJSON, &secret.CreatedAt, &secret.UpdatedAt,
			&secret.CreatedBy, &secret.UpdatedBy)
		if err != nil {
			return nil, fmt.Errorf("failed to scan SSH secret row: %w", err)
		}

		if err := json.Unmarshal(metadataJSON, &secret.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		secrets = append(secrets, secret)
	}

	return secrets, tx.Commit(context.Background())
}
