package main

import (
	"encoding/json"
	"fmt"
	handlers "kukus/nam/v2/layers/handler"
	v1 "kukus/nam/v2/layers/handler/api/rest/v1"
	"kukus/nam/v2/layers/handler/htmx"
	"log/slog"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

// Initialize and start the web server
func InitWebServer(app *Application) {
	slogLevel := slog.LevelInfo
	switch App.Configuration.WebServer.Mode {
	case "debug":
		gin.SetMode(gin.DebugMode)
		slogLevel = slog.LevelDebug
	case "release":
		gin.SetMode(gin.ReleaseMode)
	case "test":
		gin.SetMode(gin.TestMode)
		slogLevel = slog.LevelDebug
	default:
		panic("Invalid web server mode: " + App.Configuration.WebServer.Mode + ". Allowed values are: debug, release, test")
	}
	// Initialize the Gin engine with the specified mode and logging level
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slogLevel,
	}))
	log.Info("Starting web server", "mode", App.Configuration.WebServer.Mode, "port", App.Configuration.WebServer.Port)
	app.Engine = gin.New()
	app.Engine.Use(gin.Recovery())
	// Set up logging middleware
	app.Engine.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: func(params gin.LogFormatterParams) string {
			log := make(map[string]interface{})

			log["status_code"] = params.StatusCode
			log["method"] = params.Method
			log["path"] = params.Path
			log["start_time"] = params.TimeStamp.Format(time.RFC3339)
			log["remote_addr"] = params.ClientIP
			log["response_time"] = params.Latency.String()

			s, _ := json.Marshal(log)
			return string(s) + "\n"
		},
		Output: os.Stdout,
	}))
	// Set up resources
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

	app.Engine.Run(":" + fmt.Sprintf("%d", app.Configuration.WebServer.Port))
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
