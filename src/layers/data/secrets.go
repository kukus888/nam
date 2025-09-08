package data

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SecretType represents different types of secrets supported
type SecretType string

const (
	SecretTypePassword    SecretType = "password"
	SecretTypePrivateKey  SecretType = "private_key"
	SecretTypeCertificate SecretType = "certificate"
	SecretTypeAPIKey      SecretType = "api_key"
	SecretTypeSSHKey      SecretType = "ssh_key"
	SecretTypeToken       SecretType = "token"
	SecretTypeConfig      SecretType = "config"
	SecretTypeGeneric     SecretType = "generic"
)

// SecretTypeFactory creates the appropriate SecretData based on type
func SecretTypeFactory(secretType SecretType) SecretData {
	switch secretType {
	case SecretTypePassword:
		return &PasswordSecret{}
	case SecretTypePrivateKey:
		return &PrivateKeySecret{}
	case SecretTypeCertificate:
		return &CertificateSecret{}
	case SecretTypeAPIKey:
		return &APIKeySecret{}
	case SecretTypeGeneric:
		return &GenericSecret{}
	default:
		return &GenericSecret{}
	}
}

// SecretData is the interface that all secret types must implement
type SecretData interface {
	// GetType returns the type of the secret
	GetType() SecretType
	// Validate validates the secret data structure
	Validate() error
	// ToBytes converts the secret data to bytes for encryption
	ToBytes() ([]byte, error)
	// FromBytes populates the secret data from bytes after decryption
	FromBytes([]byte) error
	// GetMetadata returns additional metadata for the secret
	GetMetadata() map[string]interface{}
}

// Secret represents a secret business object
type SecretDAO struct {
	Id          uint64                 `json:"id"`
	Type        string                 `json:"type"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Data        []byte                 `json:"-"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	CreatedBy   *uint64                `json:"created_by"`
	UpdatedBy   *uint64                `json:"updated_by"`
}

// Secret represents a secret business object
type Secret struct {
	Id          uint64                 `json:"id"`
	Type        string                 `json:"type"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Data        SecretData             `json:"-"` // Excluded from JSON
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	CreatedBy   *uint64                `json:"created_by"`
	UpdatedBy   *uint64                `json:"updated_by"`
}

// SecretDTO for creating/updating secrets
type SecretDTO struct {
	Type        string                 `json:"type" binding:"required"`
	Name        string                 `json:"name" binding:"required"`
	Description string                 `json:"description"`
	Data        string                 `json:"data" binding:"required"` // The actual secret data
	Metadata    map[string]interface{} `json:"metadata"`
}

func (s *SecretDTO) ToSecretDAO(encryptedData []byte) *SecretDAO {
	secret := &SecretDAO{
		Type:        s.Type,
		Name:        s.Name,
		Description: s.Description,
		Metadata:    s.Metadata,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Data:        encryptedData,
	}

	return secret
}

func (s *SecretDTO) ToSecret(encryptedData []byte) (*Secret, error) {
	secret := &Secret{
		Type:        s.Type,
		Name:        s.Name,
		Description: s.Description,
		Metadata:    s.Metadata,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	switch SecretType(s.Type) {
	case SecretTypePassword:
		panic("Not implemented yet")
		secret.Data = &PasswordSecret{}
	case SecretTypePrivateKey:
		panic("Not implemented yet")
		secret.Data = &PrivateKeySecret{}
	case SecretTypeCertificate:
		panic("Not implemented yet")
		secret.Data = &CertificateSecret{}
	case SecretTypeAPIKey:
		panic("Not implemented yet")
		secret.Data = &APIKeySecret{}
	case SecretTypeSSHKey:
		panic("Not implemented yet")
		secret.Data = &PrivateKeySecret{}
	case SecretTypeToken:
		panic("Not implemented yet")
		secret.Data = &GenericSecret{}
	case SecretTypeConfig:
		panic("Not implemented yet")
		secret.Data = &GenericSecret{}
	case SecretTypeGeneric:
		secret.Data = &GenericSecret{
			Data: s.Data,
		}
	default:
		return nil, fmt.Errorf("unknown secret type: %s", s.Type)
	}

	return secret, nil
}

// Password secret type
type PasswordSecret struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password"`
}

func (p *PasswordSecret) GetType() SecretType {
	return SecretTypePassword
}

func (p *PasswordSecret) Validate() error {
	if p.Password == "" {
		return fmt.Errorf("password cannot be empty")
	}
	return nil
}

func (p *PasswordSecret) ToBytes() ([]byte, error) {
	return json.Marshal(p)
}

func (p *PasswordSecret) FromBytes(data []byte) error {
	return json.Unmarshal(data, p)
}

func (p *PasswordSecret) GetMetadata() map[string]interface{} {
	metadata := map[string]interface{}{
		"has_username": p.Username != "",
	}
	return metadata
}

// PrivateKey secret type
type PrivateKeySecret struct {
	KeyData    string `json:"key_data"`
	KeyType    string `json:"key_type"` // RSA, ECDSA, Ed25519
	KeySize    int    `json:"key_size"` // Key size in bits
	Passphrase string `json:"passphrase,omitempty"`
}

func (pk *PrivateKeySecret) GetType() SecretType {
	return SecretTypePrivateKey
}

func (pk *PrivateKeySecret) Validate() error {
	if pk.KeyData == "" {
		return fmt.Errorf("key data cannot be empty")
	}
	if pk.KeyType == "" {
		pk.KeyType = "RSA" // default
	}
	return nil
}

func (pk *PrivateKeySecret) ToBytes() ([]byte, error) {
	return json.Marshal(pk)
}

func (pk *PrivateKeySecret) FromBytes(data []byte) error {
	return json.Unmarshal(data, pk)
}

func (pk *PrivateKeySecret) GetMetadata() map[string]interface{} {
	return map[string]interface{}{
		"key_type":       pk.KeyType,
		"key_size":       pk.KeySize,
		"has_passphrase": pk.Passphrase != "",
	}
}

// Certificate secret type
type CertificateSecret struct {
	CertData     string    `json:"cert_data"`
	KeyData      string    `json:"key_data,omitempty"`
	CertFormat   string    `json:"cert_format"` // PEM, DER, PKCS12
	ExpiryDate   time.Time `json:"expiry_date"`
	CommonName   string    `json:"common_name"`
	Issuer       string    `json:"issuer"`
	SerialNumber string    `json:"serial_number"`
}

func (c *CertificateSecret) GetType() SecretType {
	return SecretTypeCertificate
}

func (c *CertificateSecret) Validate() error {
	if c.CertData == "" {
		return fmt.Errorf("certificate data cannot be empty")
	}
	if c.CertFormat == "" {
		c.CertFormat = "PEM" // default
	}
	return nil
}

func (c *CertificateSecret) ToBytes() ([]byte, error) {
	return json.Marshal(c)
}

func (c *CertificateSecret) FromBytes(data []byte) error {
	return json.Unmarshal(data, c)
}

func (c *CertificateSecret) GetMetadata() map[string]interface{} {
	return map[string]interface{}{
		"cert_format":     c.CertFormat,
		"expiry_date":     c.ExpiryDate,
		"common_name":     c.CommonName,
		"issuer":          c.Issuer,
		"serial_number":   c.SerialNumber,
		"has_private_key": c.KeyData != "",
	}
}

// API Key secret type
type APIKeySecret struct {
	APIKey      string            `json:"api_key"`
	APISecret   string            `json:"api_secret,omitempty"`
	ServiceName string            `json:"service_name"`
	Headers     map[string]string `json:"headers,omitempty"`
	ExpiryDate  *time.Time        `json:"expiry_date,omitempty"`
}

func (a *APIKeySecret) GetType() SecretType {
	return SecretTypeAPIKey
}

func (a *APIKeySecret) Validate() error {
	if a.APIKey == "" {
		return fmt.Errorf("API key cannot be empty")
	}
	return nil
}

func (a *APIKeySecret) ToBytes() ([]byte, error) {
	return json.Marshal(a)
}

func (a *APIKeySecret) FromBytes(data []byte) error {
	return json.Unmarshal(data, a)
}

func (a *APIKeySecret) GetMetadata() map[string]interface{} {
	metadata := map[string]interface{}{
		"service_name": a.ServiceName,
		"has_secret":   a.APISecret != "",
		"has_headers":  len(a.Headers) > 0,
	}
	if a.ExpiryDate != nil {
		metadata["expiry_date"] = a.ExpiryDate
	}
	return metadata
}

// Generic secret type for custom data
type GenericSecret struct {
	Data string `json:"data"`
}

func (g *GenericSecret) GetType() SecretType {
	return SecretTypeGeneric
}

func (g *GenericSecret) Validate() error {
	if len(g.Data) == 0 {
		return fmt.Errorf("generic secret data cannot be empty")
	}
	return nil
}

func (g *GenericSecret) ToBytes() ([]byte, error) {
	return json.Marshal(g)
}

func (g *GenericSecret) FromBytes(data []byte) error {
	return json.Unmarshal(data, g)
}

func (g *GenericSecret) GetMetadata() map[string]interface{} {
	return map[string]interface{}{
		"length": len(g.Data),
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

func GetSecretsByType(pool *pgxpool.Pool, secretType SecretType) ([]Secret, error) {
	tx, err := pool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(context.Background())

	rows, err := tx.Query(context.Background(), `
		SELECT id, type, name, description, data, metadata, created_at, updated_at, created_by, updated_by 
		FROM secrets WHERE type = $1 ORDER BY name
	`, string(secretType))
	if err != nil {
		return nil, fmt.Errorf("failed to query secrets: %w", err)
	}
	defer rows.Close()

	var secrets []Secret
	for rows.Next() {
		var secret Secret
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
