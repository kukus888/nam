package handlers

import (
	"kukus/nam/v2/layers/data"
	services "kukus/nam/v2/layers/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// SecretsHandler handles HTTP requests for secrets management
type PageSecretsHandler struct {
	database      *data.Database
	cryptoService *services.CryptoService
}

// NewPageSecretsHandler creates a new secrets handler
func NewPageSecretsHandler(database *data.Database) *PageSecretsHandler {
	return &PageSecretsHandler{
		database:      database,
		cryptoService: services.GetCryptoService(),
	}
}

// GetPageSecrets renders the secret detail page
func (h *PageSecretsHandler) GetPageSecrets(c *gin.Context) {
	secrets, err := data.GetAllSecrets(h.database.Pool)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Unable to retrieve secrets from database", "trace": err.Error()})
		return
	}
	c.HTML(http.StatusOK, "pages/secrets", gin.H{"Secrets": secrets})
}

func (h *PageSecretsHandler) GetPageEditSecret(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid secret ID"})
		return
	}
	dao, err := data.GetSecretById(h.database.Pool, id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Unable to retrieve secret from database", "trace": err.Error()})
		return
	}
	// Decrypt the secret data
	secret, err := h.cryptoService.DecryptDAO(dao)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Unable to decrypt secret data", "trace": err.Error()})
		return
	}
	dto := secret.ToDTO()
	c.HTML(http.StatusOK, "pages/secrets/edit", gin.H{"Secret": *dto})
}

func (h *PageSecretsHandler) GetPageViewSecret(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid secret ID"})
		return
	}
	dao, err := data.GetSecretById(h.database.Pool, id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Unable to retrieve secret from database", "trace": err.Error()})
		return
	}
	// Decrypt the secret data
	secret, err := h.cryptoService.DecryptDAO(dao)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Unable to decrypt secret data", "trace": err.Error()})
		return
	}
	dto := secret.ToDTO()
	c.HTML(http.StatusOK, "pages/secrets/view", gin.H{"Secret": *dto})
}
