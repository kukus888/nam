package handlers

import (
	"kukus/nam/v2/layers/data"
	services "kukus/nam/v2/layers/service"

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
		appDefDAOs, err := data.GetApplicationDefinitionsAll(pc.Database.Pool)
		if err != nil {
			ctx.AbortWithStatusJSON(500, gin.H{"error": "Unable to get healthcheck results", "trace": err.Error()})
			return
		}
		ctx.HTML(200, "pages/dashboard", gin.H{
			"AppDefDAOs": appDefDAOs,
		})
	})
	routeGroup.GET("/settings", func(ctx *gin.Context) {
		ctx.HTML(200, "pages/settings", gin.H{})
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
}
