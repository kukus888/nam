package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	handlers "kukus/nam/v2/layers/handler"
	v1 "kukus/nam/v2/layers/handler/api/rest/v1"
	"kukus/nam/v2/layers/handler/htmx"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

//go:embed web
var webResources embed.FS

//go:embed web/static
var staticResources embed.FS

// Initialize and start the web server
func InitWebServer(app *Application) {
	slogLevel := slog.LevelInfo
	switch App.Configuration.WebServer.Mode {
	case "debug", "test":
		gin.SetMode(gin.DebugMode)
		app.Engine = gin.Default()
		slogLevel = slog.LevelDebug
	case "release":
		gin.SetMode(gin.ReleaseMode)
		app.Engine = gin.New()
	default:
		panic("Invalid web server mode: " + App.Configuration.WebServer.Mode + ". Allowed values are: debug, release, test")
	}
	// Initialize the Gin engine with the specified mode and logging level
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slogLevel,
	}))
	log.Info("Starting web server", "mode", App.Configuration.WebServer.Mode, "address", App.Configuration.WebServer.Address)
	app.Engine.Use(gin.Recovery())
	// Set up logging middleware
	app.Engine.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: func(params gin.LogFormatterParams) string {
			log := make(map[string]interface{})

			log["status_code"] = params.StatusCode
			log["method"] = params.Method
			log["path"] = fmt.Sprintf("%s", params.Path) // Needed to format u/xxxx chars to readable chars
			log["start_time"] = params.TimeStamp.Format(time.RFC3339)
			log["remote_addr"] = params.ClientIP
			log["response_time"] = params.Latency.String()
			var buf bytes.Buffer
			enc := json.NewEncoder(&buf)
			enc.SetEscapeHTML(false)
			enc.Encode(log)
			return buf.String()
		},
		Output: os.Stdout,
	}))
	// Set up resources
	app.Engine.FuncMap["formatDuration"] = formatDuration
	app.Engine.FuncMap["formatTime"] = formatTime
	app.Engine.FuncMap["formatTimeRFC3339Nano"] = formatTimeRFC3339Nano
	app.Engine.FuncMap["sub1"] = sub1

	if App.Configuration.WebServer.Mode == "release" {
		LoadHTMLFromEmbedFS(app.Engine, webResources, "*")
		staticFS, err := fs.Sub(staticResources, "web/static") // Strip the web/static from the front
		if err != nil {
			panic("Failed to create sub filesystem: " + err.Error())
		}
		app.Engine.StaticFS("/static", http.FS(staticFS))
	} else {
		app.Engine.LoadHTMLGlob("./web/templates/*/*.html")
		app.Engine.Static("/static", "./web/static")
	}
	// REST
	restV1group := App.Engine.Group("/api/rest/v1")
	v1.NewApplicationController(App.Database).Init(restV1group.Group("/applications"))
	v1.NewServerController(App.Database).Init(restV1group.Group("/servers"))
	v1.NewHealthcheckController(App.Database).Init(restV1group.Group("/healthchecks"))

	// HTMX
	htmx.NewHtmxController(App.Database).Init(App.Engine.Group("/htmx"))

	// Pages
	handlers.NewPageHandler(App.Database).Init(App.Engine.Group("/"))
	handlers.NewPageSettingsHandler(App.Database).Init(App.Engine.Group("/settings"))
	handlers.NewApplicationView(App.Database).Init(App.Engine.Group("/applications"))
	handlers.NewInstanceView(App.Database).Init(App.Engine.Group("/instances"))
	handlers.NewHealthcheckView(App.Database).Init(App.Engine.Group("/healthchecks"))

	var err error
	if app.Configuration.WebServer.TLS.Enabled {
		slog.Debug("Starting web server with TLS")
		err = app.Engine.RunTLS(app.Configuration.WebServer.Address, app.Configuration.WebServer.TLS.CertPath, app.Configuration.WebServer.TLS.KeyPath)
	} else {
		err = app.Engine.Run(app.Configuration.WebServer.Address)
	}
	panic(err.Error())
}

func LoadHTMLFromEmbedFS(engine *gin.Engine, embedFS embed.FS, pattern string) {
	root := template.New("")
	tmpl := template.Must(root, LoadAndAddToRoot(engine.FuncMap, root, embedFS, pattern))
	engine.SetHTMLTemplate(tmpl)
}

// Method version
// func (engine *gin.Engine) LoadHTMLFromFS(embedFS embed.FS, pattern string) {
// 	root := template.New("")
// 	tmpl := template.Must(root, LoadAndAddToRoot(engine.FuncMap, root, embedFS, pattern))
// 	engine.SetHTMLTemplate(tmpl)
// }

func LoadAndAddToRoot(funcMap template.FuncMap, rootTemplate *template.Template, embedFS embed.FS, pattern string) error {
	pattern = strings.ReplaceAll(pattern, ".", "\\.")
	pattern = strings.ReplaceAll(pattern, "*", ".*")

	err := fs.WalkDir(embedFS, ".", func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if matched, _ := regexp.MatchString(pattern, path); !d.IsDir() && matched {
			data, readErr := embedFS.ReadFile(path)
			if readErr != nil {
				return readErr
			}
			t := rootTemplate.New(path).Funcs(funcMap)
			if _, parseErr := t.Parse(string(data)); parseErr != nil {
				return parseErr
			}
		}
		return nil
	})
	return err
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

func formatTimeRFC3339Nano(t time.Time) string {
	return t.Format(time.RFC3339Nano)
}

func sub1(x int) int {
	return x - 1
}
