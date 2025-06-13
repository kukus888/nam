package htmx

import (
	"kukus/nam/v2/layers/data"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ItemView struct {
	Database *data.Database
}

func NewItemView(database *data.Database) ItemView {
	return ItemView{
		Database: database,
	}
}

func (iv ItemView) Init(routeGroup *gin.RouterGroup) {
	routeGroup.GET("/servers/view", iv.ServerView)
	routeGroup.GET("/servers/update/:id", iv.ServerUpdate)
	routeGroup.GET("/servers/create", iv.ServerCreate)
}

func (tnc ItemView) GetItemContainer(ctx *gin.Context) {
	itemType := ctx.Param("type")
	itemUrl := "/api/rest/v1/" + itemType
	ctx.HTML(200, "template/items/itemcontainer", gin.H{"itemUrl": itemUrl})
}

func (tnc ItemView) ServerView(ctx *gin.Context) {
	servers, err := data.GetServerAll(tnc.Database.Pool)
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}
	ctx.HTML(200, "template/items/server/view", gin.H{"Servers": servers})
}

func (tnc ItemView) ServerUpdate(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}
	dao, err := data.GetServerById(tnc.Database.Pool, uint(id))
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}
	ctx.HTML(200, "template/items/server/update", gin.H{"Server": &dao})
}

func (tnc ItemView) ServerCreate(ctx *gin.Context) {
	ctx.HTML(200, "template/items/server/create", nil)
}
