package htmx

import (
	"kukus/nam/v2/layers/data"
	services "kukus/nam/v2/layers/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ApplicationInstanceView struct {
	Service *services.ApplicationInstanceService
}

func NewApplicationInstanceView(database *data.Database) *ApplicationInstanceView {
	return &ApplicationInstanceView{
		Service: &services.ApplicationInstanceService{
			Database: database,
		},
	}
}

func (aiv ApplicationInstanceView) Init(routeGroup *gin.RouterGroup) {
	routeGroup.GET("/:id", aiv.RenderApplicationInstanceSmall)
}

func (aiv *ApplicationInstanceView) RenderApplicationInstanceSmall(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.AbortWithStatusJSON(400, err)
		return
	}
	instance, err := aiv.Service.GetApplicationInstanceById(id)
	if err != nil {
		ctx.AbortWithStatusJSON(500, err)
		return
	}
	ctx.HTML(200, "template/application.small", gin.H{
		"ApplicationInstance": *instance,
	})
}
