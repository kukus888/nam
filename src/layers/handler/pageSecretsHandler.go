package handlers

import (
	"kukus/nam/v2/layers/data"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SecretsHandler handles HTTP requests for secrets management
type PageSecretsHandler struct {
	database *data.Database
}

// NewPageSecretsHandler creates a new secrets handler
func NewPageSecretsHandler(database *data.Database) *PageSecretsHandler {
	return &PageSecretsHandler{
		database: database,
	}
}

// GetPageSecrets renders the secret detail page
func (h *PageSecretsHandler) GetPageSecrets(c *gin.Context) {
	secrets, err := data.GetAllSecrets(h.database.Pool)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Unable to retrieve secrets", "trace": err.Error()})
		return
	}
	c.HTML(http.StatusOK, "pages/secrets", gin.H{"Secrets": secrets})
}
