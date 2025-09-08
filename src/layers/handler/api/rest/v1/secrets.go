package v1

import (
	"net/http"
	"strconv"

	"kukus/nam/v2/layers/data"
	services "kukus/nam/v2/layers/service"

	"github.com/gin-gonic/gin"
)

// SecretsHandler handles HTTP requests for secrets management
type SecretsHandler struct {
	secretsService *services.SecretsService
}

// NewSecretsHandler creates a new secrets handler
func NewSecretsHandler(secretsService *services.SecretsService) *SecretsHandler {
	return &SecretsHandler{
		secretsService: secretsService,
	}
}

// CreateSecret creates a new secret
func (h *SecretsHandler) CreateSecret(c *gin.Context) {
	var dto data.SecretDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format", "trace": err.Error()})
		return
	}
	// Get user ID from context (assuming it's set by auth middleware)
	userIdInterface, exists := c.Get("user_id")
	var userId *uint64
	if exists {
		if uid, ok := userIdInterface.(uint64); ok {
			userId = &uid
		}
	}

	id, err := h.secretsService.CreateSecret(&dto, userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create secret", "trace": err.Error()})
		return
	}

	c.Header("HX-Redirect", "/secrets")
	c.JSON(http.StatusCreated, gin.H{"id": *id, "message": "Secret created successfully"})
}

// GetSecret retrieves a secret by ID (returns metadata only, not the actual secret data)
func (h *SecretsHandler) GetSecret(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid secret ID"})
		return
	}

	secret, _, err := h.secretsService.GetSecret(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Secret not found", "trace": err.Error()})
		return
	}

	// Return metadata only for security
	secret.Data = nil
	c.JSON(http.StatusOK, secret)
}

// GetSecretData retrieves the decrypted secret data (sensitive operation)
func (h *SecretsHandler) GetSecretData(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid secret ID"})
		return
	}

	secret, secretData, err := h.secretsService.GetSecret(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Secret not found", "trace": err.Error()})
		return
	}

	response := gin.H{
		"id":          secret.Id,
		"name":        secret.Name,
		"type":        secret.Type,
		"description": secret.Description,
		"metadata":    secret.Metadata,
		"data":        secretData,
	}

	c.JSON(http.StatusOK, response)
}

// UpdateSecret updates an existing secret
func (h *SecretsHandler) UpdateSecret(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid secret ID"})
		return
	}

	var dto data.SecretDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format", "trace": err.Error()})
		return
	}
	// Get user ID from context
	userIdInterface, exists := c.Get("user_id")
	var userId *uint64
	if exists {
		if uid, ok := userIdInterface.(uint64); ok {
			userId = &uid
		}
	}

	if err := h.secretsService.UpdateSecret(id, &dto, userId); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update secret", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Secret updated successfully"})
}

// DeleteSecret deletes a secret
func (h *SecretsHandler) DeleteSecret(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid secret ID"})
		return
	}

	if err := h.secretsService.DeleteSecret(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete secret", "details": err.Error()})
		return
	}

	c.Header("HX-Redirect", "/secrets")
	c.JSON(http.StatusOK, gin.H{"message": "Secret deleted successfully"})
}

// GetSecretTypes returns the available secret types
func (h *SecretsHandler) GetSecretTypes(c *gin.Context) {
	types := []gin.H{
		{"type": string(data.SecretTypePassword), "description": "Username/password combinations"},
		{"type": string(data.SecretTypePrivateKey), "description": "Private keys (RSA, ECDSA, Ed25519)"},
		{"type": string(data.SecretTypeCertificate), "description": "SSL/TLS certificates with optional private keys"},
		{"type": string(data.SecretTypeAPIKey), "description": "API keys and tokens"},
		{"type": string(data.SecretTypeSSHKey), "description": "SSH private/public key pairs"},
		{"type": string(data.SecretTypeToken), "description": "Authentication tokens"},
		{"type": string(data.SecretTypeConfig), "description": "Configuration files with sensitive data"},
		{"type": string(data.SecretTypeGeneric), "description": "Generic key-value secret data"},
	}

	c.JSON(http.StatusOK, gin.H{"types": types})
}
