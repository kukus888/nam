package handlers

import (
	"kukus/nam/v2/layers/data"
	services "kukus/nam/v2/layers/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PageController struct {
	Database        *data.Database
	ServerService   services.ServerService
	TopologyService services.TopologyNodeService
	ItemService     services.ItemService
}

func NewPageController(database *data.Database) PageController {
	return PageController{
		Database:        database,
		ServerService:   services.ServerService{Database: database},
		TopologyService: services.TopologyNodeService{Database: database},
		ItemService:     services.ItemService{Database: database},
	}
}

func (pc PageController) Init(routeGroup *gin.RouterGroup) {
	routeGroup.GET("/", func(ctx *gin.Context) {
		// TODO: Health page?
		ctx.HTML(200, "pages/index", gin.H{})
	})
	routeGroup.GET("/nodes", func(ctx *gin.Context) {
		nodes, err := pc.TopologyService.GetAllTopologyNodes()
		if err != nil {
			ctx.AbortWithStatus(500)
			return
		}
		ctx.HTML(200, "pages/nodes", gin.H{
			"Nodes": nodes,
		})
	})
	routeGroup.GET("/items", func(ctx *gin.Context) {
		ctx.HTML(200, "pages/items", gin.H{
			"Types": pc.ItemService.GetAllItemTypes(),
		})
	})
	routeGroup.GET("/servers", func(ctx *gin.Context) {
		servers, err := data.GetServerAll(pc.Database.Pool)
		if err != nil {
			ctx.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
			return
		}
		ctx.HTML(200, "pages/servers", gin.H{
			"Servers": servers,
		})
	})
	routeGroup.GET("/applications", func(ctx *gin.Context) {
		apps, err := data.GetApplicationDefinitions(pc.Database.Pool)
		if err != nil {
			ctx.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
			return
		}
		ctx.HTML(200, "pages/applications", gin.H{
			"Applications": apps,
		})
	})
	routeGroup.GET("/applications/create", func(ctx *gin.Context) {
		hcs, err := data.GetHealthChecksAll(pc.Database.Pool)
		if err != nil {
			ctx.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
			return
		}
		servers, err := data.GetServerAll(pc.Database.Pool)
		if err != nil {
			ctx.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
			return
		}
		ctx.HTML(200, "pages/applications/create", gin.H{
			"Healthchecks": &hcs,
			"Servers":      &servers,
		})
	})
	routeGroup.GET("/applications/:id/details", func(ctx *gin.Context) {
		id, err := strconv.Atoi(ctx.Param("id"))
		if err != nil {
			ctx.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
			return
		}
		app, err := data.GetApplicationDefinitionById(pc.Database.Pool, uint64(id))
		if err != nil {
			ctx.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
			return
		}
		var hc *data.Healthcheck
		if app.HealthcheckId != nil {
			hc, err = data.GetHealthCheckById(pc.Database.Pool, uint(*app.HealthcheckId))
			if err != nil {
				ctx.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
				return
			}
		}
		instances, err := app.GetInstances(pc.Database.Pool)
		if err != nil {
			ctx.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
			return
		}
		ctx.HTML(200, "pages/applications/details", gin.H{
			"Application": app,
			"Healthcheck": hc,
			"Instances":   instances,
		})
	})
}
