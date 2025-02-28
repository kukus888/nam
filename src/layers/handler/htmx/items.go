package htmx

import (
	"kukus/nam/v2/layers/data"
	services "kukus/nam/v2/layers/service"

	"github.com/gin-gonic/gin"
)

type ItemView struct {
	Database    *data.Database
	ItemService services.ItemService
}

func NewItemView(database *data.Database) ItemView {
	return ItemView{
		Database:    database,
		ItemService: services.NewItemService(database),
	}
}

func (iv ItemView) Init(routeGroup *gin.RouterGroup) {
	routeGroup.GET("/types", iv.GetItemTypes)
	routeGroup.DELETE("/new/:itemType", iv.NewItemInstance)
	routeGroup.GET("/container", iv.GetItemContainer)
}

func (tnc ItemView) GetItemTypes(ctx *gin.Context) {
	types := []string{
		"application_instance",
		"application_definition",
		"server",
		"topology_node",
	}
	ctx.JSON(200, gin.H{"types": types})
}

func (tnc ItemView) NewItemInstance(ctx *gin.Context) {
	types := []string{
		"application_instance",
		"application_definition",
		"server",
		"topology_node",
	}
	ctx.JSON(200, gin.H{"types": types})
}

func (tnc ItemView) GetItemContainer(ctx *gin.Context) {
	itemType := ctx.Param("type")
	itemUrl := "/api/rest/v1/" + itemType
	ctx.HTML(200, "template/items/itemcontainer", gin.H{"itemUrl": itemUrl})
}
