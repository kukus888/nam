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
	c.HTML(http.StatusOK, "pages/servers/create", gin.H{})
}

func (h PageServerHandler) GetPageServerEdit(c *gin.Context) {
	serverID := c.Param("id")
	id, err := strconv.Atoi(serverID)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "Invalid server ID", "trace": err.Error()})
		return
	}
	server, err := data.GetServerById(h.Database.Pool, uint(id))
	if err != nil {
		c.AbortWithStatusJSON(404, gin.H{"error": "Server not found", "trace": err.Error()})
		return
	}
	c.HTML(http.StatusOK, "pages/servers/edit", gin.H{
		"Server": server,
	})
}
