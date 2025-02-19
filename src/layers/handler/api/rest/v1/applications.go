package applications

import (
	data "kukus/nam/v2/layers/data"

	"github.com/gin-gonic/gin"
)

type ApplicationController struct {
	Database *data.Database
}

// Initializes new Controller on declared RouterGroup, with specified resources
func (ac *ApplicationController) Init(routerGroup *gin.RouterGroup) {
	routerGroup.POST("/", ac.Post)
	routerGroup.GET("/", ac.Get)
	routerGroup.PATCH("/", ac.Patch)
	routerGroup.PUT("/", ac.Put)
	routerGroup.DELETE("/", ac.Get)
}

func (ac *ApplicationController) Get(ctx *gin.Context) {
	ctx.AbortWithStatus(500)
}

func (ac *ApplicationController) Post(ctx *gin.Context) {
	ctx.AbortWithStatus(500)
}

func (ac *ApplicationController) Patch(ctx *gin.Context) {
	ctx.AbortWithStatus(500)
}

func (ac *ApplicationController) Put(ctx *gin.Context) {
	ctx.AbortWithStatus(500)
}

func (ac *ApplicationController) Delete(ctx *gin.Context) {
	ctx.AbortWithStatus(500)
}
