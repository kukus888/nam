package handlers

import (
	"kukus/nam/v2/layers/data"
	"strconv"

	"github.com/gin-gonic/gin"
)

type HealthcheckView struct {
	Database *data.Database
}

func NewHealthcheckView(database *data.Database) HealthcheckView {
	return HealthcheckView{
		Database: database,
	}
}

func (av HealthcheckView) Init(routeGroup *gin.RouterGroup) {
	routeGroup.GET("/", func(ctx *gin.Context) {
		hcs, err := data.GetHealthChecksAll(av.Database.Pool)
		if err != nil {
			ctx.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
			return
		}
		ctx.HTML(200, "pages/healthchecks", gin.H{
			"Healthchecks": hcs,
		})
	})
	routeGroup.GET("/create", func(ctx *gin.Context) {
		hcs, err := data.GetHealthChecksAll(av.Database.Pool)
		if err != nil {
			ctx.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
			return
		}
		ctx.HTML(200, "pages/healthchecks/create", gin.H{
			"Healthchecks": &hcs,
		})
	})
	idGroup := routeGroup.Group("/:id")
	{
		idGroup.GET("/details", av.GetPageHealthcheckDetails)
	}
}

func (av HealthcheckView) GetPageHealthcheckDetails(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}
	hc, err := data.GetHealthCheckById(av.Database.Pool, uint(id))
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}
	ctx.HTML(200, "pages/healthchecks/details", gin.H{
		"Healthcheck": hc,
	})
}

func (av HealthcheckView) GetPageHealthcheckInstanceNew(ctx *gin.Context) {
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
