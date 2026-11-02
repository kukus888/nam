package handlers

import (
	"fmt"
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
	Id          int
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
			Id:          i,
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

func (pc PageSettingsHandler) PostTimerEnable(ctx *gin.Context) {
	id := ctx.Param("id")
	timerId, err := strconv.Atoi(id)
	if err != nil {
		ctx.HTML(400, "pages/settings/timers", gin.H{"error": "Invalid timer ID", "trace": err.Error()})
		return
	}
	ts := services.GetTimerService()
	job, found := ts.Jobs[timerId]
	if !found {
		ctx.HTML(400, "pages/settings/timers", gin.H{"error": "Timer not found"})
		return
	}
	job.Enable()
	ctx.Header("HX-Redirect", "/settings/timers")
	ctx.Status(200)
}

func (pc PageSettingsHandler) PostTimerDisable(ctx *gin.Context) {
	id := ctx.Param("id")
	timerId, err := strconv.Atoi(id)
	if err != nil {
		ctx.HTML(400, "pages/settings/timers", gin.H{"error": "Invalid timer ID", "trace": err.Error()})
		return
	}
	ts := services.GetTimerService()
	job, found := ts.Jobs[timerId]
	if !found {
		ctx.HTML(400, "pages/settings/timers", gin.H{"error": "Timer not found"})
		return
	}
	job.Disable()
	ctx.Header("HX-Redirect", "/settings/timers")
	ctx.Status(200)
}

func (pc PageSettingsHandler) PostTimerRun(ctx *gin.Context) {
	id := ctx.Param("id")
	timerId, err := strconv.Atoi(id)
	if err != nil {
		ctx.HTML(400, "pages/settings/timers", gin.H{"error": "Invalid timer ID", "trace": err.Error()})
		return
	}
	ts := services.GetTimerService()
	job, found := ts.Jobs[timerId]
	if !found {
		ctx.HTML(400, "pages/settings/timers", gin.H{"error": "Timer not found"})
		return
	}
	job.Run()
	ctx.Status(200)
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

func (pc PageSettingsHandler) GetPageConfigurationSettings(ctx *gin.Context) {
	settings, err := data.GetAllAppSettings(pc.Database.Pool)
	if err != nil {
		ctx.HTML(500, "pages/settings/configuration", gin.H{"error": "Unable to load configuration settings", "trace": err.Error()})
		return
	}

	dbSize, err := pc.Database.GetDatabaseSize()
	if err != nil {
		ctx.HTML(500, "pages/settings/configuration", gin.H{"error": "Unable to read current database size", "trace": err.Error()})
		return
	}

	guardCfg, cfgErr := data.GetDatabaseCapacityGuardConfig(pc.Database.Pool)
	usagePercent := 0.0
	if cfgErr == nil && guardCfg.AllocatedBytes > 0 {
		usagePercent = (float64(*dbSize) / float64(guardCfg.AllocatedBytes)) * 100
	}

	resultsCount, countErr := data.GetHealthcheckResultsCount(pc.Database.Pool)
	if countErr != nil {
		ctx.HTML(500, "pages/settings/configuration", gin.H{"error": "Unable to count healthcheck results", "trace": countErr.Error()})
		return
	}

	ctx.HTML(200, "pages/settings/configuration", gin.H{
		"AppSettings":              settings,
		"CurrentDatabaseSizeBytes": *dbSize,
		"HealthcheckResultsCount":  *resultsCount,
		"CapacityUsagePercent":     usagePercent,
	})
}

func (pc PageSettingsHandler) PostConfigurationSettings(ctx *gin.Context) {
	settings, err := data.GetAllAppSettings(pc.Database.Pool)
	if err != nil {
		ctx.HTML(500, "pages/settings/configuration", gin.H{"error": "Unable to load configuration settings", "trace": err.Error()})
		return
	}
	overrides := make(map[string]string, len(*settings))
	capacityOverrides := make(map[string]string)
	for _, setting := range *settings {
		formField := "value_" + setting.Key
		value := strings.TrimSpace(ctx.PostForm(formField))
		overrides[setting.Key] = value
		if strings.HasPrefix(setting.Key, "db_capacity_") {
			capacityOverrides[setting.Key] = value
		}
	}

	if validateErr := data.ValidateDatabaseCapacityGuardSettingsCandidate(pc.Database.Pool, capacityOverrides); validateErr != nil {
		dbSize, _ := pc.Database.GetDatabaseSize()
		resultsCount, _ := data.GetHealthcheckResultsCount(pc.Database.Pool)
		currentDbSize := int64(0)
		if dbSize != nil {
			currentDbSize = *dbSize
		}
		currentResultsCount := int64(0)
		if resultsCount != nil {
			currentResultsCount = *resultsCount
		}
		ctx.HTML(400, "pages/settings/configuration", gin.H{
			"error":                    "Invalid capacity settings",
			"trace":                    validateErr.Error(),
			"AppSettings":              settings,
			"CurrentDatabaseSizeBytes": currentDbSize,
			"HealthcheckResultsCount":  currentResultsCount,
			"CapacityUsagePercent":     0.0,
		})
		return
	}

	userID := ctx.GetUint64("user_id")
	for _, setting := range *settings {
		newValue := overrides[setting.Key]
		if updateErr := data.UpdateAppSettingValue(pc.Database.Pool, setting.Key, newValue, &userID); updateErr != nil {
			dbSize, _ := pc.Database.GetDatabaseSize()
			resultsCount, _ := data.GetHealthcheckResultsCount(pc.Database.Pool)
			currentDbSize := int64(0)
			if dbSize != nil {
				currentDbSize = *dbSize
			}
			currentResultsCount := int64(0)
			if resultsCount != nil {
				currentResultsCount = *resultsCount
			}
			ctx.HTML(400, "pages/settings/configuration", gin.H{
				"error":                    fmt.Sprintf("Invalid value for %s", setting.Key),
				"trace":                    updateErr.Error(),
				"AppSettings":              settings,
				"CurrentDatabaseSizeBytes": currentDbSize,
				"HealthcheckResultsCount":  currentResultsCount,
				"CapacityUsagePercent":     0.0,
			})
			return
		}
	}

	if guardCfg, cfgErr := data.GetDatabaseCapacityGuardConfig(pc.Database.Pool); cfgErr == nil {
		ts := services.GetTimerService()
		if job, found := ts.Jobs[2]; found {
			if timer, ok := job.(*services.DatabaseCapacityGuardTimer); ok {
				timer.SetInterval(guardCfg.CheckInterval)
			}
		}
	}

	ctx.Redirect(303, "/settings/configuration")
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
