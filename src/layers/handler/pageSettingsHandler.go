package handlers

import (
	"kukus/nam/v2/layers/data"
	"strconv"
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

func (pc PageSettingsHandler) GetPageUsers(ctx *gin.Context) {
	usersFull, err := data.GetAllUsersFull(pc.Database.Pool)
	if err != nil {
		ctx.HTML(500, "pages/settings/users", gin.H{"error": "Unable to get user list", "trace": err.Error()})
		return
	}
	ctx.HTML(200, "pages/settings/users", gin.H{"users": usersFull})
}

func (pc PageSettingsHandler) GetPageUserCreate(ctx *gin.Context) {
	roles, err := data.GetAllRoles(pc.Database.Pool)
	if err != nil {
		ctx.HTML(500, "pages/settings/users/create", gin.H{"error": "Unable to get role list", "trace": err.Error()})
		return
	}
	ctx.HTML(200, "pages/settings/users/create", gin.H{"roles": roles})
}

func (pc PageSettingsHandler) GetPageUserEdit(ctx *gin.Context) {
	id := ctx.Param("id")
	userId, err := strconv.Atoi(id)
	if err != nil {
		ctx.HTML(400, "pages/settings/users/edit", gin.H{"error": "Invalid user ID", "trace": err.Error()})
		return
	}
	user, err := data.GetUserById(pc.Database.Pool, uint64(userId))
	if err != nil {
		ctx.HTML(500, "pages/settings/users/edit", gin.H{"error": "Unable to get user", "trace": err.Error()})
		return
	}
	roles, err := data.GetAllRoles(pc.Database.Pool)
	if err != nil {
		ctx.HTML(500, "pages/settings/users/edit", gin.H{"error": "Unable to get role list", "trace": err.Error()})
		return
	}
	ctx.HTML(200, "pages/settings/users/edit", gin.H{"user": user, "roles": roles})
}
