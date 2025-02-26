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
		servers, err := pc.ServerService.GetAllServers()
		if err != nil {
			ctx.AbortWithStatus(500)
			return
		}
		ctx.HTML(200, "pages/index", gin.H{
			"Servers": servers,
		})
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
}
