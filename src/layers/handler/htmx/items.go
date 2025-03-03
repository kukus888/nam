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
	routeGroup.GET("/view/servers", iv.GetServerView)
}

func (tnc ItemView) GetItemContainer(ctx *gin.Context) {
	itemType := ctx.Param("type")
	itemUrl := "/api/rest/v1/" + itemType
	ctx.HTML(200, "template/items/itemcontainer", gin.H{"itemUrl": itemUrl})
}

func (tnc ItemView) GetServerView(ctx *gin.Context) {
	servers, err := data.GetServerAll(tnc.Database.Pool)
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": err})
		return
	}
	ctx.HTML(200, "template/items/server/view", gin.H{"Servers": servers})
}
