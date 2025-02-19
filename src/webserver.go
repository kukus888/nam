package main

import (
	applications "kukus/nam/v2/layers/handler/api/rest/v1"

	"github.com/gin-gonic/gin"
)

// Initialize and start the web server
func InitWebServer(app *Application) {
	app.Engine = gin.Default()
	app.Engine.LoadHTMLGlob("./web/templates/*")
	app.Engine.Static("/static", "./web/static")
	app.Engine.GET("/", func(c *gin.Context) {
		c.HTML(200, "pages/index", gin.H{})
	})
	restV1group := App.Engine.Group("/api/rest/v1")
	applicationRouteGroup := restV1group.Group("/applications")
	applicationController := applications.ApplicationController{Database: App.Database}
	applicationController.Init(applicationRouteGroup)

}
