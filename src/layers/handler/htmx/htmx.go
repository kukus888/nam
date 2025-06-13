package htmx

import (
	"kukus/nam/v2/layers/data"

	"github.com/gin-gonic/gin"
)

type HtmxController struct {
	Database *data.Database
}

func NewHtmxController(database *data.Database) HtmxController {
	return HtmxController{
		Database: database,
	}
}

func (hc HtmxController) Init(routeGroup *gin.RouterGroup) {
	NewApplicationView(hc.Database).Init(routeGroup.Group("/applications"))
	NewHtmxHealthHandler(hc.Database.Pool).Init(routeGroup.Group("/health"))

	NewItemView(hc.Database).Init(routeGroup.Group("/items"))

}
