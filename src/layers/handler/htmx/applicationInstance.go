package htmx

import (
	"kukus/nam/v2/layers/data"
	services "kukus/nam/v2/layers/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ApplicationView struct {
	Service *services.ApplicationInstanceService
}

func NewApplicationView(database *data.Database) *ApplicationView {
	return &ApplicationView{
		Service: &services.ApplicationInstanceService{
			Database: database,
		},
	}
}

func (aiv ApplicationView) Init(routeGroup *gin.RouterGroup) {
	applicationGroup := routeGroup.Group("/:appId")
	{
		applicationGroup.GET("/instances/:instanceId/small", aiv.RenderApplicationInstanceSmall)
	}
}

func (aiv *ApplicationView) RenderApplicationInstanceSmall(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("instanceId"), 10, 64)
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
