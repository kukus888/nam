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

/*
 *	Component used for viewing Applications (Definitions and instances), their components, and pages
 */
func (av ApplicationView) Init(routeGroup *gin.RouterGroup) {
	routeGroup.GET("/", func(ctx *gin.Context) {
		apps, err := data.GetApplicationDefinitions(av.Database.Pool)
		if err != nil {
			ctx.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
			return
		}
		ctx.HTML(200, "pages/applications", gin.H{
			"Applications": apps,
		})
	})
	routeGroup.GET("/create", func(ctx *gin.Context) {
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
	})
	idGroup := routeGroup.Group("/:id")
	{
		idGroup.GET("/details", av.GetPageApplicationDetails)
		idGroup.GET("/instances/create", av.GetPageApplicationInstanceNew)
		idGroup.GET("/instances/:instanceId/details", av.GetPageApplicationInstanceDetails)
	}
}

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

func (av ApplicationView) GetPageApplicationInstanceDetails(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}
	appDefinition, err := data.GetApplicationDefinitionById(av.Database.Pool, uint64(id))
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}
	instanceId, err := strconv.Atoi(ctx.Param("instanceId"))
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}
	appInstance, err := data.GetApplicationInstanceById(av.Database.Pool, uint64(instanceId))
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	} else if appInstance == nil {
		ctx.AbortWithStatusJSON(404, gin.H{"error": "Application instance not found"})
		return
	}
	// Check if the instance belongs to the application
	if appInstance.ApplicationDefinitionID != appDefinition.Id {
		ctx.AbortWithStatusJSON(418, gin.H{"error": "Application instance does not belong to this application"})
		return
	}
	var hc *data.Healthcheck
	if appDefinition.HealthcheckId != nil {
		hc, err = data.GetHealthCheckById(av.Database.Pool, uint(*appDefinition.HealthcheckId))
		if err != nil {
			ctx.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
			return
		}
	}
	ctx.HTML(200, "pages/application/instance/details", gin.H{
		"Application": appDefinition,
		"Healthcheck": hc,
		"Instance":    appInstance,
	})
}

func (av ApplicationView) GetPageApplicationInstanceNew(ctx *gin.Context) {
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
