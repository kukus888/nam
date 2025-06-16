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
		apps, err := data.GetApplicationDefinitionsAll(av.Database.Pool)
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
	// route /applications/:id
	idGroup := routeGroup.Group("/:id")
	{
		idGroup.GET("/edit", av.GetPageApplicationEdit)
		idGroup.GET("/details", av.GetPageApplicationDetails)
		idGroup.GET("/instances/create", av.GetPageApplicationInstanceNew)
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

// Renders a page "New Application Instance"
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

// Renders a page "Edit Application Definition"
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
