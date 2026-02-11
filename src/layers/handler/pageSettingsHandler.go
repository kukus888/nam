package handlers

import (
	"kukus/nam/v2/layers/data"
	services "kukus/nam/v2/layers/service"
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

// TimerJobView is a simplified representation of a TimerJob, designed for display purposes in the settings page. It includes human-readable fields for the timer's name, description, enabled status, and interval.
type TimerJobView struct {
	Name        string
	Description string
	Enabled     bool
	Interval    string
}

// GetPageTimerSettings handles the GET request for the timer settings page. It retrieves the list of timer jobs from the TimerService, converts them into a view-friendly format, and renders the HTML template with the timer data.
func (pc PageSettingsHandler) GetPageTimerSettings(ctx *gin.Context) {
	ts := services.GetTimerService()
	timerViews := make([]TimerJobView, len(ts.Jobs))
	for i, job := range ts.Jobs {
		timerViews[i] = TimerJobView{
			Name:        job.GetName(),
			Description: job.GetDescription(),
			Enabled:     job.IsEnabled(),
			Interval:    job.GetInterval().String(),
		}
	}
	ctx.HTML(200, "pages/settings/timers", gin.H{
		"Timers": timerViews,
	})
}

func (pc PageSettingsHandler) GetPageDatabaseSettings(ctx *gin.Context) {
	connConfig := pc.Database.Pool.Config().ConnConfig
	connStrSafe := strings.ReplaceAll(connConfig.ConnString(), connConfig.Password, "****")
	tableSizes, err := data.GetTableSizes(pc.Database.Pool)
	if err != nil {
		ctx.HTML(500, "pages/settings/database", gin.H{"error": "Unable to get table sizes", "trace": err.Error()})
		return
	}
	ctx.HTML(200, "pages/settings/database", gin.H{
		"DbConfigConnString":    connStrSafe,
		"DbConfigRuntimeParams": connConfig.RuntimeParams,
		"TableSizes":            tableSizes,
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
