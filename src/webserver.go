package main

import (
	handlers "kukus/nam/v2/layers/handler"
	v1 "kukus/nam/v2/layers/handler/api/rest/v1"
	"kukus/nam/v2/layers/handler/htmx"

	"github.com/gin-gonic/gin"
)

// Initialize and start the web server
func InitWebServer(app *Application) {
	app.Engine = gin.Default()
	app.Engine.LoadHTMLGlob("./web/templates/*/*.html")
	app.Engine.Static("/static", "./web/static")
	// REST
	restV1group := App.Engine.Group("/api/rest/v1")
	v1.NewApplicationController(App.Database).Init(restV1group.Group("/applications"))
	v1.NewServerController(App.Database).Init(restV1group.Group("/servers"))

	// HTMX
	htmx.NewHtmxController(App.Database).Init(App.Engine.Group("/htmx"))

	// Pages
	handlers.NewPageController(App.Database).Init(App.Engine.Group("/"))
	app.Engine.Run(":8080")
	// TODO: debug | prod
	// TODO: logging
}
