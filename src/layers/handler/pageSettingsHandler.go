package handlers

import (
	"kukus/nam/v2/layers/data"
	"strings"

	"github.com/gin-gonic/gin"
)

type PageSettingsHandler struct {
	Database *data.Database
}

func NewPageSettingsHandler(database *data.Database) PageSettingsHandler {
	return PageSettingsHandler{
		Database: database,
	}
}

func (pc PageSettingsHandler) GetPageSettings(ctx *gin.Context) {
	ctx.HTML(200, "pages/settings", gin.H{})
}

func (pc PageSettingsHandler) GetPageDatabaseSettings(ctx *gin.Context) {
	connConfig := pc.Database.Pool.Config().ConnConfig
	connStrSafe := strings.ReplaceAll(connConfig.ConnString(), connConfig.Password, "****")
	ctx.HTML(200, "pages/settings/database", gin.H{
		"DbConfigConnString":    connStrSafe,
		"DbConfigRuntimeParams": connConfig.RuntimeParams,
	})
}
