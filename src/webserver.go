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
		app.Engine.Use(ErrorHandler())
		// Set default jwt key for debugging
		services.SetJWTKey([]byte("nam-debug-jwt-key-2025"))
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
	app.Engine.FuncMap["derefBool"] = derefBool
	app.Engine.FuncMap["derefInt"] = derefInt
	app.Engine.FuncMap["derefInt64"] = derefInt64
	app.Engine.FuncMap["derefUint64"] = derefUint64
	app.Engine.FuncMap["derefStr"] = derefStr
	app.Engine.FuncMap["title"] = func(s string) string {
		if len(s) == 0 {
			return s
		}
		return strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
	}
	app.Engine.FuncMap["lower"] = strings.ToLower
	app.Engine.FuncMap["printf"] = fmt.Sprintf
	app.Engine.FuncMap["sub"] = func(a, b int) int { return a - b }

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
		restV1group.Use(RequireRole(dbPool, "Viewer")) // All endpoint require at least Viewer role
		{                                              // Servers
			serverController := apiRestV1.NewServerController(App.Database)
			serverGroup := restV1group.Group("/servers")
			serverGroup.POST("/", RequireRole(dbPool, "Operator"), serverController.NewServer)
			serverGroup.GET("/", serverController.GetAll)
			serverIdGroup := serverGroup.Group("/:serverId")
			{ // Server ID specific routes
				serverIdGroup.GET("/", serverController.GetById)
				serverIdGroup.PUT("/", RequireRole(dbPool, "Operator"), serverController.UpdateById)
				serverIdGroup.DELETE("/", RequireRole(dbPool, "Operator"), serverController.RemoveById)
			}
		}
		{ // Healthchecks
			hcController := apiRestV1.NewHealthcheckController(App.Database)
			hcGroup := restV1group.Group("/healthchecks")
			hcGroup.POST("/", RequireRole(dbPool, "Operator"), hcController.NewHealthcheck)
			hcGroup.GET("/", hcController.GetAll)
			hcIdGroup := hcGroup.Group("/:hcId")
			{ // Healthcheck ID specific routes
				hcIdGroup.GET("/", hcController.GetById)
				hcIdGroup.PUT("/", RequireRole(dbPool, "Operator"), hcController.UpdateHealthcheck)
				hcIdGroup.DELETE("/", RequireRole(dbPool, "Operator"), hcController.Delete)
			}
		}
		{ // Application Definitions
			appDefController := apiRestV1.NewApplicationDefinitionController(App.Database)
			appDefGroup := restV1group.Group("/applications")
			appDefGroup.POST("/", RequireRole(dbPool, "Operator"), appDefController.NewApplication)
			appDefGroup.GET("/", appDefController.GetAll)
			appDefIdGroup := appDefGroup.Group("/:appId")
			{ // Application ID specific routes
				appDefIdGroup.GET("/", appDefController.GetById)
				appDefIdGroup.PUT("/", RequireRole(dbPool, "Operator"), appDefController.UpdateApplicationDefinition)
				appDefIdGroup.DELETE("/", RequireRole(dbPool, "Operator"), appDefController.DeleteById)
				{ // Application Definition Variables
					appDefVarController := apiRestV1.NewAppDefVariablesController(App.Database)
					appDefVarGroup := appDefIdGroup.Group("/variables")
					appDefVarGroup.POST("/", RequireRole(dbPool, "Operator"), appDefVarController.CreateVariable)
					{
						appDefVarIdGroup := appDefVarGroup.Group("/:varId")
						appDefVarIdGroup.PUT("/", RequireRole(dbPool, "Operator"), appDefVarController.UpdateVariable)
						appDefVarIdGroup.DELETE("/", RequireRole(dbPool, "Operator"), appDefVarController.DeleteVariable)
					}
				}
				{ // Application Instances
					appInsGroup := appDefIdGroup.Group("/instances")
					appInsController := apiRestV1.NewApplicationInstanceController(App.Database)
					appInsGroup.POST("/", RequireRole(dbPool, "Operator"), appInsController.CreateInstance)
					appInsGroup.GET("/", appInsController.GetAllInstances)
					appInsIdGroup := appInsGroup.Group("/:instanceId")
					{ // Instance ID specific routes
						appInsIdGroup.GET("/", appInsController.GetById)
						appInsIdGroup.DELETE("/", RequireRole(dbPool, "Operator"), appInsController.DeleteInstance)
						appInsIdGroup.POST("/maintenance", RequireRole(dbPool, "Operator"), appInsController.ToggleMaintenance)
						{ // Application Instance Variables
							appInsVarsController := apiRestV1.NewAppInstanceVariablesController(App.Database)
							appInsVarsGroup := appInsIdGroup.Group("/variables")
							appInsVarsGroup.GET("/", appInsVarsController.GetAllVariables)
							appInsVarsGroup.POST("/", RequireRole(dbPool, "Operator"), appInsVarsController.CreateVariable)
							appInsVarIdGroup := appInsVarsGroup.Group("/:varId")
							appInsVarIdGroup.PUT("/", RequireRole(dbPool, "Operator"), appInsVarsController.UpdateVariable)
							appInsVarIdGroup.DELETE("/", RequireRole(dbPool, "Operator"), appInsVarsController.DeleteVariable)
						}
					}
				}
			}
		}
		{ // Users
			userHandler := apiRestV1.NewUserHandler(dbPool)
			userGroup := restV1group.Group("/users")
			userGroup.Use(RequireRole(dbPool, "Admin")) // Only admin can manage users
			userGroup.POST("/create", userHandler.CreateUser)
			userGroup.DELETE("/:id", userHandler.DeleteUser)
			userGroup.PUT("/:id", userHandler.UpdateUser)
			userGroup.PUT("/:id/password", userHandler.UpdatePassword)
		}
		{ // Secrets
			secretService := services.NewSecretsService(App.Database.Pool, log, cryptoService)
			secretHandler := apiRestV1.NewSecretsHandler(secretService)
			secretGroup := restV1group.Group("/secrets")
			secretGroup.Use(RequireRole(dbPool, "Operator"))       // Only admin and operator can manage secrets
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
		{ // Actions
			actionController := apiRestV1.NewActionController(App.Database)
			actionGroup := restV1group.Group("/actions")

			// Action Templates
			actionTemplatesGroup := actionGroup.Group("/templates")
			actionTemplatesGroup.GET("/", actionController.GetAllActionTemplates)
			actionTemplatesGroup.POST("/", RequireRole(dbPool, "Operator"), actionController.CreateActionTemplate)
			actionTemplateIdGroup := actionTemplatesGroup.Group("/:templateId")
			{
				actionTemplateIdGroup.GET("/", actionController.GetActionTemplateById)
				actionTemplateIdGroup.PUT("/", RequireRole(dbPool, "Operator"), actionController.UpdateActionTemplate)
				actionTemplateIdGroup.DELETE("/", RequireRole(dbPool, "Operator"), actionController.DeleteActionTemplate)
			}

			// Actions
			actionGroup.GET("/", actionController.GetAllActions)
			actionGroup.POST("/", RequireRole(dbPool, "Operator"), actionController.CreateAction)
			actionGroup.POST("/preflight", actionController.PreflightCheck)

			actionIdGroup := actionGroup.Group("/:actionId")
			{
				actionIdGroup.GET("/", actionController.GetActionById)
				actionIdGroup.POST("/start", RequireRole(dbPool, "Operator"), actionController.StartAction)
				actionIdGroup.POST("/cancel", RequireRole(dbPool, "Operator"), actionController.CancelAction)
				actionIdGroup.GET("/status", actionController.GetActionStatus)
			}

			// Execution logs
			restV1group.GET("/executions/:executionId/logs", actionController.GetExecutionLogs)
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
	// Dashboard - accessible to all authenticated users (Viewer and above)
	rootGroup.GET("/", RequireRole(dbPool, "Viewer"), ph.GetPageDashboard)
	rootGroup.GET("/dashboard", RequireRole(dbPool, "Viewer"), ph.GetPageDashboard)
	rootGroup.GET("/dashboard/component", RequireRole(dbPool, "Viewer"), ph.GetDashboardComponent)
	rootGroup.GET("/dashboard/data", RequireRole(dbPool, "Viewer"), ph.GetDashboardDataAPI)
	rootGroup.GET("/profile", ph.GetProfilePage)
	{ // Servers
		psh := handlers.NewPageServerHandler(App.Database)
		// Server viewing - accessible to Viewers and above
		rootGroup.GET("/servers", RequireRole(dbPool, "Viewer"), psh.GetPageServers)
		rootGroup.GET("/servers/:id/view", RequireRole(dbPool, "Viewer"), psh.GetPageServerView)
		// Server management - accessible to Operators and above
		rootGroup.GET("/servers/create", RequireRole(dbPool, "Operator"), psh.GetPageServerCreate)
		rootGroup.GET("/servers/:id/edit", RequireRole(dbPool, "Operator"), psh.GetPageServerEdit)
	}
	{ // Application Definitions
		av := handlers.NewApplicationView(App.Database)
		// Application viewing - accessible to Viewers and above
		rootGroup.GET("/applications", RequireRole(dbPool, "Viewer"), av.GetPageApplications)
		rootGroup.GET("/applications/:id/details", RequireRole(dbPool, "Viewer"), av.GetPageApplicationDetails)
		// Application management - accessible to Operators and above
		rootGroup.GET("/applications/create", RequireRole(dbPool, "Operator"), av.GetPageApplicationCreate)
		rootGroup.GET("/applications/maintenance", RequireRole(dbPool, "Operator"), av.GetPageApplicationMaintenance)
		rootGroup.GET("/applications/:id/edit", RequireRole(dbPool, "Operator"), av.GetPageApplicationEdit)
		rootGroup.GET("/applications/:id/variables", RequireRole(dbPool, "Operator"), av.GetPageApplicationVariables)
		rootGroup.GET("/applications/:id/instances/create", RequireRole(dbPool, "Operator"), av.GetPageApplicationInstanceCreate)
	}
	{ // Application instances
		iv := handlers.NewInstanceView(App.Database)
		// Instance viewing - accessible to Viewers and above
		rootGroup.GET("/instances/:id/details", RequireRole(dbPool, "Viewer"), iv.GetPageApplicationInstanceDetails)
		// Instance management - accessible to Operators and above
		rootGroup.GET("/instances/:id/variables", RequireRole(dbPool, "Operator"), iv.GetPageApplicationInstanceVariables)
	}
	{ // Healthchecks
		hcv := handlers.NewHealthcheckView(App.Database)
		// Healthcheck viewing - accessible to Viewers and above
		rootGroup.GET("/healthchecks", RequireRole(dbPool, "Viewer"), hcv.GetPageHealthchecks)
		rootGroup.GET("/healthchecks/:id/details", RequireRole(dbPool, "Viewer"), hcv.GetPageHealthcheckDetails)
		// Healthcheck management - accessible to Operators and above
		rootGroup.GET("/healthchecks/create", RequireRole(dbPool, "Operator"), hcv.GetPageHealthcheckCreate)
		rootGroup.GET("/healthchecks/:id/edit", RequireRole(dbPool, "Operator"), hcv.GetPageHealthcheckEdit)
	}
	{ // Settings
		psh := handlers.NewPageSettingsHandler(App.Database)
		routeGroup := rootGroup.Group("/settings")
		routeGroup.Use(RequireRole(dbPool, "Admin"))
		routeGroup.GET("/", psh.GetPageSettings)
		routeGroup.GET("/database", psh.GetPageDatabaseSettings)
		routeGroup.GET("/users", psh.GetPageUsers)
		routeGroup.GET("/users/create", psh.GetPageUserCreate) // Placeholder for user creation page
		routeGroup.GET("/users/:id/edit", psh.GetPageUserEdit)
	}
	{ // Secrets Management
		psh := handlers.NewPageSecretsHandler(App.Database, cryptoService)
		secretGroup := rootGroup.Group("/secrets")
		secretGroup.Use(RequireRole(dbPool, "Operator")) // Only admin and operator can manage secrets
		secretGroup.GET("/", psh.GetPageSecrets)
		secretGroup.GET("/:id/edit", psh.GetPageEditSecret)
		secretGroup.GET("/:id/details", psh.GetPageViewSecret)
	}
	{ // Actions
		av := handlers.NewActionView(App.Database)
		routeGroup := rootGroup.Group("/actions")
		routeGroup.Use(RequireRole(dbPool, "Viewer")) // All authenticated users can view actions

		routeGroup.GET("/", av.GetPageActions)
		routeGroup.GET("/new", av.GetPageActionCreate)
		routeGroup.POST("/preflight", RequireRole(dbPool, "Operator"), av.PostActionsPreflight) // HTMX preflight endpoint

		// Action Templates
		routeGroup.GET("/templates", av.GetPageActionTemplates)
		routeGroup.GET("/templates/new", av.GetPageActionTemplateCreate)
		routeGroup.GET("/templates/:id/edit", av.GetPageActionTemplateEdit)
		routeGroup.POST("/templates/:id/edit", RequireRole(dbPool, "Operator"), av.PostPageActionTemplateEdit)
		routeGroup.GET("/templates/:id/details", av.GetPageActionTemplateDetails)

		// Action Details
		routeGroup.GET("/:id/details", av.GetPageActionDetails)
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

// ErrorHandler captures errors and returns a consistent JSON error response
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next() // Step1: Process the request first.

		// Step2: Check if any errors were added to the context
		if len(c.Errors) > 0 {
			// Step3: Use the last error
			err := c.Errors.Last().Err

			// Step4: Respond with a generic error message
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": err.Error(),
			})
		}

		// Any other steps if no errors are found
	}
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
func derefBool(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

// Dereference a pointer to an integer value
func derefInt(i *int) int {
	if i == nil {
		return 0
	}
	return *i
}

func derefInt64(i *int64) int64 {
	if i == nil {
		return 0
	}
	return *i
}

func derefUint64(i *uint64) uint64 {
	if i == nil {
		return 0
	}
	return *i
}

// Dereference a pointer to a string value
func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
