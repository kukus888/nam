package handlers

import (
	applications "kukus/nam/v2/layers/handler/api/rest/v1"

	"github.com/gin-gonic/gin"
)

func RegisterHandlers() {
	g := gin.Default()
	restV1group := g.Group("/api/rest/v1")
	applicationGroup := restV1group.Group("/applications")
	applicationGroup.POST("/", applications.Post)
	applicationGroup.GET("/", applications.Get)
	applicationGroup.PATCH("/", applications.Patch)
	applicationGroup.PUT("/", applications.Put)
	applicationGroup.DELETE("/", applications.Get)
}
