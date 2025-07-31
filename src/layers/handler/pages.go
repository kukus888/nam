package handlers

import (
	"kukus/nam/v2/layers/data"

	"github.com/gin-gonic/gin"
)

type PageHandler struct {
	Database *data.Database
}

func NewPageHandler(database *data.Database) PageHandler {
	return PageHandler{
		Database: database,
	}
}

func (pc PageHandler) GetPageDashboard(ctx *gin.Context) {
	appDefDAOs, err := data.GetApplicationDefinitionsAll(pc.Database.Pool)
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": "Unable to get application definitions", "trace": err.Error()})
		return
	}
	ctx.HTML(200, "pages/dashboard", gin.H{
		"AppDefDAOs": appDefDAOs,
	})
}

func (pc PageHandler) GetPageServers(ctx *gin.Context) {
	servers, err := data.GetServerAll(pc.Database.Pool)
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": "Unable to get servers", "trace": err.Error()})
		return
	}
	ctx.HTML(200, "pages/servers", gin.H{
		"Servers": servers,
	})
}
