package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	data "kukus/nam/v2/layers/data"
	handlers "kukus/nam/v2/layers/handler"
	apiRestV1 "kukus/nam/v2/layers/handler/api/rest/v1"
	"kukus/nam/v2/layers/handler/htmx"
	services "kukus/nam/v2/layers/service"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
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
		app.Engine = gin.New()
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
	app.Engine.NoRoute(handlers.NotFound)
	// Set up resources
	app.Engine.FuncMap["formatDuration"] = formatDuration
	app.Engine.FuncMap["formatTime"] = formatTime
	app.Engine.FuncMap["formatTimeRFC3339Nano"] = formatTimeRFC3339Nano
	app.Engine.FuncMap["sub1"] = sub1
	app.Engine.FuncMap["add"] = func(a, b int) int { return a + b }
	app.Engine.FuncMap["deref"] = deref

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
	// Alias to shorten code
	dbPool := App.Database.Pool
	cryptoService := services.NewCryptoService("nam-secrets-salt-2025", []byte("nam-secrets-salt-2025"))
	{ // REST
		restV1group := App.Engine.Group("/api/rest/v1")
		restV1group.Use(AuthMiddleware())
		{ // Servers
			serverController := apiRestV1.NewServerController(App.Database)
			serverGroup := restV1group.Group("/servers")
			serverGroup.Use(RequireRole(dbPool, "admin"))
			serverGroup.POST("/", serverController.NewServer)
			serverGroup.GET("/", serverController.GetAll)
			serverIdGroup := serverGroup.Group("/:serverId")
			{ // Server ID specific routes
				serverIdGroup.GET("/", serverController.GetById)
				serverIdGroup.PUT("/", serverController.UpdateById)
				serverIdGroup.DELETE("/", serverController.RemoveById)
			}
		}
		{ // Healthchecks
			hcController := apiRestV1.NewHealthcheckController(App.Database)
			hcGroup := restV1group.Group("/healthchecks")
			hcGroup.Use(RequireRole(dbPool, "admin"))
			hcGroup.POST("/", hcController.NewHealthcheck)
			hcGroup.GET("/", hcController.GetAll)
			hcIdGroup := hcGroup.Group("/:hcId")
			{ // Healthcheck ID specific routes
				hcIdGroup.GET("/", hcController.GetById)
				hcIdGroup.PUT("/", hcController.UpdateHealthcheck)
				hcIdGroup.DELETE("/", hcController.Delete)
			}
		}
		{ // Application Definitions
			appDefController := apiRestV1.NewApplicationDefinitionController(App.Database)
			appDefGroup := restV1group.Group("/applications")
			appDefGroup.Use(RequireRole(dbPool, "admin"))
			appDefGroup.POST("/", appDefController.NewApplication)
			appDefGroup.GET("/", appDefController.GetAll)
			appDefIdGroup := appDefGroup.Group("/:appId")
			{ // Application ID specific routes
				appDefIdGroup.GET("/", appDefController.GetById)
				appDefIdGroup.PUT("/", appDefController.UpdateApplicationDefinition)
				appDefIdGroup.DELETE("/", appDefController.DeleteById)
				{ // Application Definition Variables
					appDefVarController := apiRestV1.NewAppDefVariablesController(App.Database)
					appDefVarGroup := appDefIdGroup.Group("/variables")
					appDefVarGroup.POST("/", appDefVarController.CreateVariable)
					{
						appDefVarIdGroup := appDefVarGroup.Group("/:varId")
						appDefVarIdGroup.PUT("/", appDefVarController.UpdateVariable)
						appDefVarIdGroup.DELETE("/", appDefVarController.DeleteVariable)
					}
				}
				{ // Application Instances
					appInsGroup := appDefIdGroup.Group("/instances")
					appInsController := apiRestV1.NewApplicationInstanceController(App.Database)
					appInsGroup.POST("/", appInsController.CreateInstance)
					appInsGroup.GET("/", appInsController.GetAllInstances)
					appInsIdGroup := appInsGroup.Group("/:instanceId")
					{ // Instance ID specific routes
						appInsIdGroup.GET("/", appInsController.GetById)
						appInsIdGroup.DELETE("/", appInsController.DeleteInstance)
						appInsIdGroup.POST("/maintenance", appInsController.ToggleMaintenance)
						{ // Application Instance Variables
							appInsVarsController := apiRestV1.NewAppInstanceVariablesController(App.Database)
							appInsVarsGroup := appInsIdGroup.Group("/variables")
							appInsVarsGroup.GET("/", appInsVarsController.GetAllVariables)
							appInsVarsGroup.POST("/", appInsVarsController.CreateVariable)
							appInsVarIdGroup := appInsVarsGroup.Group("/:varId")
							appInsVarIdGroup.PUT("/", appInsVarsController.UpdateVariable)
							appInsVarIdGroup.DELETE("/", appInsVarsController.DeleteVariable)
						}
					}
				}
			}
		}
		{ // Users
			userHandler := apiRestV1.NewUserHandler(dbPool)
			userGroup := restV1group.Group("/users")
			userGroup.Use(RequireRole(dbPool, "admin")) // Only admin can manage users
			userGroup.POST("/create", userHandler.CreateUser)
			userGroup.DELETE("/:id", userHandler.DeleteUser)
			userGroup.PUT("/:id", userHandler.UpdateUser)
			userGroup.PUT("/:id/password", userHandler.UpdatePassword)
		}
		{ // Secrets
			secretService := services.NewSecretsService(App.Database.Pool, log, cryptoService)
			secretHandler := apiRestV1.NewSecretsHandler(secretService)
			secretGroup := restV1group.Group("/secrets")
			secretGroup.Use(RequireRole(dbPool, "admin"))
			secretGroup.POST("/", secretHandler.CreateSecret)      // Create new secret
			secretGroup.PUT("/:id", secretHandler.UpdateSecret)    // Update secret
			secretGroup.DELETE("/:id", secretHandler.DeleteSecret) // Delete secret
		}
		{ // Profile
			profileHandler := apiRestV1.NewProfileHandler(dbPool)
			profileGroup := restV1group.Group("/profile")
			profileGroup.PUT("/", profileHandler.UpdateUser)
			profileGroup.PUT("/password", profileHandler.UpdatePassword)
		}
	}
	{ // HTMX
		htmxGroup := App.Engine.Group("/htmx")
		htmxGroup.Use(AuthMiddleware())
		htmx.NewHtmxController(App.Database).Init(htmxGroup)
	}

	// Pages
	handlers.NewLoginPageHandler(App.Database).Init(App.Engine.Group("/login"))
	// Handlers for the main pages, protected by authentication middleware
	rootGroup := App.Engine.Group("/")
	rootGroup.Use(AuthMiddleware())

	ph := handlers.NewPageHandler(App.Database)
	rootGroup.GET("/", RequireRole(dbPool, "admin"), ph.GetPageDashboard)
	rootGroup.GET("/dashboard", RequireRole(dbPool, "admin"), ph.GetPageDashboard)
	rootGroup.GET("/dashboard/component", RequireRole(dbPool, "admin"), ph.GetDashboardComponent)
	rootGroup.GET("/dashboard/data", RequireRole(dbPool, "admin"), ph.GetDashboardDataAPI)
	rootGroup.GET("/profile", ph.GetProfilePage)
	{ // Servers
		psh := handlers.NewPageServerHandler(App.Database)
		rootGroup.GET("/servers", RequireRole(dbPool, "admin"), psh.GetPageServers)
		rootGroup.GET("/servers/create", RequireRole(dbPool, "admin"), psh.GetPageServerCreate)
		rootGroup.GET("/servers/:id/edit", RequireRole(dbPool, "admin"), psh.GetPageServerEdit)
	}
	{ // Application Definitions
		av := handlers.NewApplicationView(App.Database)
		routeGroup := rootGroup.Group("/applications")
		routeGroup.Use(RequireRole(dbPool, "admin"))
		routeGroup.GET("/", av.GetPageApplications)
		routeGroup.GET("/create", av.GetPageApplicationCreate)
		routeGroup.GET("/maintenance", av.GetPageApplicationMaintenance)
		idGroup := routeGroup.Group("/:id")
		{ // Application ID specific routes
			idGroup.GET("/details", av.GetPageApplicationDetails)
			idGroup.GET("/edit", av.GetPageApplicationEdit)
			idGroup.GET("/variables", av.GetPageApplicationVariables)
			idGroup.GET("/instances/create", av.GetPageApplicationInstanceCreate)
		}
	}
	{ // Application instances
		iv := handlers.NewInstanceView(App.Database)
		routeGroup := rootGroup.Group("/instances")
		routeGroup.Use(RequireRole(dbPool, "admin"))
		idGroup := routeGroup.Group("/:id")
		{ // Instance ID specific routes
			idGroup.GET("/details", iv.GetPageApplicationInstanceDetails)
			idGroup.GET("/variables", iv.GetPageApplicationInstanceVariables)
		}
	}
	{ // Healthchecks
		hcv := handlers.NewHealthcheckView(App.Database)
		routeGroup := rootGroup.Group("/healthchecks")
		routeGroup.Use(RequireRole(dbPool, "admin"))
		routeGroup.GET("/", hcv.GetPageHealthchecks)
		routeGroup.GET("/create", hcv.GetPageHealthcheckCreate)
		idGroup := routeGroup.Group("/:id")
		{ // Healthcheck ID specific routes
			idGroup.GET("/details", hcv.GetPageHealthcheckDetails)
			idGroup.GET("/edit", hcv.GetPageHealthcheckEdit)
		}
	}
	{ // Settings
		psh := handlers.NewPageSettingsHandler(App.Database)
		routeGroup := rootGroup.Group("/settings")
		routeGroup.Use(RequireRole(dbPool, "admin"))
		routeGroup.GET("/", psh.GetPageSettings)
		routeGroup.GET("/database", psh.GetPageDatabaseSettings)
		routeGroup.GET("/users", psh.GetPageUsers)
		routeGroup.GET("/users/create", psh.GetPageUserCreate) // Placeholder for user creation page
		routeGroup.GET("/users/:id/edit", psh.GetPageUserEdit)
	}
	{ // Secrets Management
		psh := handlers.NewPageSecretsHandler(App.Database, cryptoService)
		secretGroup := rootGroup.Group("/secrets")
		secretGroup.Use(RequireRole(dbPool, "admin"))
		secretGroup.GET("/", psh.GetPageSecrets)
		secretGroup.GET("/:id/edit", psh.GetPageEditSecret)
		secretGroup.GET("/:id/details", psh.GetPageViewSecret)
	}

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

// Middleware to check JWT token and set user context
// This middleware should be used for routes that require authentication
// Gets the token from the Authorization header, or from a cookie if not present
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			cookie, err := c.Cookie("token")
			if err == nil {
				tokenString = cookie
			} else {
				handlers.Unauthorized(c) // No token provided
				return
			}
		}

		claims := &services.Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
			return services.GetJWTKeyProvider().Key, nil
		})

		if err != nil || !token.Valid {
			handlers.Unauthorized(c) // Invalid token
			return
		}

		// Check if the token is expiring soon (5 minutes)
		if time.Until(claims.ExpiresAt.Time) < 5*time.Minute {
			newTokenString, err := services.RegenerateToken(claims)
			if err == nil {
				// Set the new token in the cookie
				c.SetCookie("token", newTokenString, 3600, "/", "", false, true)
			}
		}

		// Set user information in the context
		c.Set("username", claims.Username)
		c.Set("user_id", claims.UserID)

		c.Next()
	}
}

// RBAC Middleware to check user roles for RBAC
// db - the database connection pool
// requiredRole - the role that the user must have to access the route
// If the user does not have the required role, a 403 Forbidden response is returned
func RequireRole(db *pgxpool.Pool, requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		username, exists := c.Get("username")
		if !exists {
			handlers.Unauthorized(c)
			return
		}

		user, err := data.GetUserByUsername(db, username.(string))
		if err != nil || user == nil {
			handlers.Unauthorized(c)
			return
		}

		if !user.HasRole(requiredRole) {
			handlers.Forbidden(c)
			return
		}

		c.Next()
	}
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

// Dereference a pointer to a boolean value
func deref(b *bool) bool {
	return *b
}
