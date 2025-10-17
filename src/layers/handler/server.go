package handlers

import (
	"kukus/nam/v2/layers/data"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PageServerHandler struct {
	Database *data.Database
}

func NewPageServerHandler(database *data.Database) PageServerHandler {
	return PageServerHandler{
		Database: database,
	}
}

func (h PageServerHandler) GetPageServers(ctx *gin.Context) {
	servers, err := data.GetServerAll(h.Database.Pool)
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": "Unable to get servers", "trace": err.Error()})
		return
	}
	ctx.HTML(200, "pages/servers", gin.H{
		"Servers": servers,
	})
}

func (h PageServerHandler) GetPageServerCreate(c *gin.Context) {
	// Get SSH-compatible secrets for dropdown
	secrets, err := data.GetSshSecrets(h.Database.Pool)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Unable to get secrets", "trace": err.Error()})
		return
	}
	c.HTML(http.StatusOK, "pages/servers/create", gin.H{
		"SshSecrets": secrets,
	})
}

func (h PageServerHandler) GetPageServerEdit(c *gin.Context) {
	serverID := c.Param("id")
	id, err := strconv.ParseUint(serverID, 10, 64)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "Invalid server ID", "trace": err.Error()})
		return
	}
	server, err := data.GetServerById(h.Database.Pool, id)
	if err != nil {
		c.AbortWithStatusJSON(404, gin.H{"error": "Server not found", "trace": err.Error()})
		return
	}
	// Get SSH-compatible secrets for dropdown
	secrets, err := data.GetSshSecrets(h.Database.Pool)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Unable to get secrets", "trace": err.Error()})
		return
	}
	c.HTML(http.StatusOK, "pages/servers/edit", gin.H{
		"Server":     server,
		"SshSecrets": secrets,
	})
}

func (h PageServerHandler) GetPageServerView(c *gin.Context) {
	serverID := c.Param("id")
	id, err := strconv.ParseUint(serverID, 10, 64)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "Invalid server ID", "trace": err.Error()})
		return
	}
	server, err := data.GetServerById(h.Database.Pool, id)
	if err != nil {
		c.AbortWithStatusJSON(404, gin.H{"error": "Server not found", "trace": err.Error()})
		return
	}

	// Get the secret name if SSH is configured
	var sshSecretName string
	if server.SshAuthSecretId != nil {
		secret, err := data.GetSecretById(h.Database.Pool, *server.SshAuthSecretId)
		if err == nil && secret != nil {
			sshSecretName = secret.Name
		}
	}

	c.HTML(http.StatusOK, "pages/servers/view", gin.H{
		"Server":        server,
		"SshSecretName": sshSecretName,
	})
}
