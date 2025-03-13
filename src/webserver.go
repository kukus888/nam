package main

import (
	"fmt"
	handlers "kukus/nam/v2/layers/handler"
	v1 "kukus/nam/v2/layers/handler/api/rest/v1"
	"kukus/nam/v2/layers/handler/htmx"
	"time"

	"github.com/gin-gonic/gin"
)

// Initialize and start the web server
func InitWebServer(app *Application) {
	app.Engine = gin.Default()
	app.Engine.FuncMap["formatDuration"] = formatDuration
	app.Engine.FuncMap["formatTime"] = formatTime
	app.Engine.LoadHTMLGlob("./web/templates/*/*.html")
	app.Engine.Static("/static", "./web/static")
	// REST
	restV1group := App.Engine.Group("/api/rest/v1")
	v1.NewApplicationController(App.Database).Init(restV1group.Group("/applications"))
	v1.NewServerController(App.Database).Init(restV1group.Group("/servers"))
	v1.NewHealthcheckController(App.Database).Init(restV1group.Group("/healthchecks"))

	// HTMX
	htmx.NewHtmxController(App.Database).Init(App.Engine.Group("/htmx"))

	// Pages
	handlers.NewPageController(App.Database).Init(App.Engine.Group("/"))
	handlers.NewApplicationView(App.Database).Init(App.Engine.Group("/applications"))
	handlers.NewHealthcheckView(App.Database).Init(App.Engine.Group("/healthchecks"))
	app.Engine.Run(":8080")
	// TODO: debug | prod
	// TODO: logging
}

// Add these to your template functions
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	} else if d < time.Minute {
		return fmt.Sprintf("%.1fs", float64(d)/float64(time.Second))
	} else if d < time.Hour {
		return fmt.Sprintf("%.1fm", float64(d)/float64(time.Minute))
	} else {
		return fmt.Sprintf("%.1fh", float64(d)/float64(time.Hour))
	}
}

func formatTime(t time.Time) string {
	return t.Format("Jan 02, 15:04:05")
}
