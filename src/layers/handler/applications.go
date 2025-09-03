package handlers

import (
	"kukus/nam/v2/layers/data"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ApplicationView struct {
	Database *data.Database
}

func NewApplicationView(database *data.Database) ApplicationView {
	return ApplicationView{
		Database: database,
	}
}

// GetPageApplications renders the page with a list of all applications
func (av ApplicationView) GetPageApplications(ctx *gin.Context) {
	apps, err := data.GetApplicationDefinitionsAll(av.Database.Pool)
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}
	ctx.HTML(200, "pages/applications", gin.H{
		"Applications": apps,
	})
}

// GetPageApplicationCreate renders the page to create a new application definition
func (av ApplicationView) GetPageApplicationCreate(ctx *gin.Context) {
	hcs, err := data.GetHealthChecksAll(av.Database.Pool)
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}
	servers, err := data.GetServerAll(av.Database.Pool)
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}
	ctx.HTML(200, "pages/applications/create", gin.H{
		"Healthchecks": &hcs,
		"Servers":      &servers,
	})
}

// GetPageApplicationInstanceCreate renders the page to create a new application instance
func (av ApplicationView) GetPageApplicationInstanceCreate(ctx *gin.Context) {
	appId, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}
	app, err := data.GetApplicationDefinitionById(av.Database.Pool, uint64(appId))
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}
	servers, err := data.GetServerAll(av.Database.Pool)
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}
	ctx.HTML(200, "pages/applications/instances/create", gin.H{
		"Application": app,
		"Servers":     servers,
	})
}

// GetPageApplicationDetails renders the page to view details of an application definition
func (av ApplicationView) GetPageApplicationDetails(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}
	app, err := data.GetApplicationDefinitionById(av.Database.Pool, uint64(id))
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}
	var hc *data.Healthcheck
	if app.HealthcheckId != nil {
		hc, err = data.GetHealthCheckById(av.Database.Pool, uint(*app.HealthcheckId))
		if err != nil {
			ctx.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
			return
		}
	}
	instances, err := data.GetApplicationInstancesFullByApplicationDefinitionId(av.Database.Pool, uint64(app.Id))
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}
	ctx.HTML(200, "pages/applications/details", gin.H{
		"Application": app,
		"Healthcheck": hc,
		"Instances":   instances,
	})
}

// Render the page to edit an application definition
func (av ApplicationView) GetPageApplicationEdit(ctx *gin.Context) {
	appId, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{"error": "Reqest must contain application definition id", "trace": err.Error()})
		return
	}
	app, err := data.GetApplicationDefinitionById(av.Database.Pool, uint64(appId))
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": "Unable to get application definition id", "trace": err.Error()})
		return
	}
	hcs, err := data.GetHealthChecksAll(av.Database.Pool)
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": "Unable to get health checks", "trace": err.Error()})
		return
	}
	ctx.HTML(200, "pages/applications/edit", gin.H{
		"Application":  app,
		"Healthchecks": hcs,
	})
}

func (av ApplicationView) GetPageApplicationMaintenance(ctx *gin.Context) {
	instances, err := data.GetAllApplicationInstancesFull(av.Database.Pool)
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}
	ctx.HTML(200, "pages/applications/maintenance", gin.H{
		"Instances": instances,
	})
}
