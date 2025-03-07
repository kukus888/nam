package htmx

import (
	"kukus/nam/v2/layers/data"
	services "kukus/nam/v2/layers/service"

	"github.com/gin-gonic/gin"
)

type HtmxController struct {
	Database      *data.Database
	ServerService services.ServerService
}

func NewHtmxController(database *data.Database) HtmxController {
	return HtmxController{
		Database: database,
		ServerService: services.ServerService{
			Database: database,
		},
	}
}

func (hc HtmxController) Init(routeGroup *gin.RouterGroup) {
	routeGroup.GET("/servers/:id", hc.Server)
	NewHealthcheckView(hc.Database).Init(routeGroup.Group("/healthchecks"))
	NewApplicationView(hc.Database).Init(routeGroup.Group("/applications"))

	NewTopologyNodeController(hc.Database).Init(routeGroup.Group("/nodes"))
	NewItemView(hc.Database).Init(routeGroup.Group("/items"))

}
