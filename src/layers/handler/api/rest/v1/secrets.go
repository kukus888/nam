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

	c.Header("HX-Redirect", "/secrets/"+idStr+"/details")
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
