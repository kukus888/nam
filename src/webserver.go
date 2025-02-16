package main

import (
	"math/rand"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Initialize and start the web server
func InitWebServer() {
	engine := gin.Default()
	engine.LoadHTMLGlob("./web/templates/*")
	engine.Static("/static", "./web/static")
	engine.GET("/", func(c *gin.Context) {
		c.HTML(200, "pages/index", gin.H{})
	})
	engine.GET("/applications", func(c *gin.Context) {
		// TODO: Cache
		appDefs, err := DbQueryTypeAll(ApplicationDefinition{})
		if err != nil {
			c.AbortWithStatus(500)
			return
		}
		c.HTML(200, "pages/applications", gin.H{
			"appDefs": appDefs,
		})
	})
	restGroup := engine.Group("/api/rest")
	{
		applicationGroup := restGroup.Group("/applications")
		{
			applicationGroup.GET("/", func(c *gin.Context) { // View All ApplicationDefinition
				apps, err := DbQueryTypeAll(ApplicationDefinition{})
				if err != nil {
					c.JSON(500, gin.H{"error": "Unable to get ApplicationDefinition", "trace": err})
					return
				}
				c.JSON(200, apps)
			})
			applicationGroup.POST("/", func(c *gin.Context) { // Create ApplicationDefinition
				var appDef ApplicationDefinition
				if err := c.ShouldBindJSON(&appDef); err != nil {
					c.JSON(400, gin.H{"error": "Invalid JSON", "trace": err})
					return
				}
				if err := appDef.DbInsert(); err != nil {
					c.JSON(500, gin.H{"error": "Unable to save ApplicationDefinition", "trace": err})
					return
				}
				c.JSON(201, appDef)
			})
			applicationIdGroup := applicationGroup.Group("/:appId")
			{
				applicationIdGroup.GET("/", func(c *gin.Context) { // View ApplicationDefinition by ID
					appId := c.Param("appId")
					app, err := DbQueryTypeSingleWithParams(ApplicationDefinition{}, DbFilter{
						Column:   "id",
						Operator: DbOperatorEqual,
						Value:    appId,
					})
					if err != nil {
						c.JSON(500, gin.H{"error": "Unable to get ApplicationDefinition", "trace": err})
						return
					} else {
						c.JSON(200, app)
					}
				})
				applicationIdGroup.POST("/instances", func(c *gin.Context) { // Create ApplicationInstance
					var appInsDto ApplicationInstanceDTO
					if err := c.ShouldBindJSON(&appInsDto); err != nil {
						c.JSON(400, gin.H{"error": "Invalid JSON", "trace": err})
						return
					}
					err := appInsDto.DbInsert()
					if err != nil {
						c.JSON(500, gin.H{"error": "Unable to insert application instances", "trace": err})
						return
					} else {
						c.JSON(201, appInsDto)
					}
				})
				applicationIdGroup.GET("/instances", func(c *gin.Context) { // View All ApplicationDefinition ApplicationInstance-s
					appId := c.Param("appId")
					inst, err := DbQueryTypeSingleWithParams(ApplicationInstance{}, DbFilter{
						Column:   "application_definition_id",
						Operator: DbOperatorEqual,
						Value:    appId,
					})
					if err != nil { // TODO: FIXXX!!
						c.JSON(500, gin.H{"error": "Unable to get application instances", "trace": err})
						return
					} else {
						c.JSON(200, inst)
					}
				})
			}

		}
		serverGroup := restGroup.Group("/servers")
		{
			serverGroup.GET("/", func(c *gin.Context) {
				servers, err := DbQueryTypeAll(Server{})
				if err != nil {
					c.JSON(500, gin.H{"error": "Unable to get Server list from DB", "trace": err})
				} else {
					c.JSON(200, servers)
				}
			})
			serverGroup.POST("/", func(c *gin.Context) {
				var server Server
				if err := c.ShouldBindJSON(&server); err != nil {
					c.JSON(400, gin.H{"error": "Invalid JSON", "trace": err})
					return
				}
				err := server.DbInsert()
				if err != nil {
					c.JSON(500, gin.H{"error": "Unable to insert server", "trace": err})
				} else {
					c.JSON(201, server)
				}
			})
			serverGroup.GET("/:serverId", func(c *gin.Context) {
				serverId := c.Param("serverId")
				id, err := strconv.Atoi(serverId)
				if err != nil {
					c.JSON(400, gin.H{"error": "Server ID not string"})
				}
				def := Server{ID: uint(id)}
				if err != nil {
					c.JSON(500, gin.H{"error": "Unable to get application instances", "trace": err})
				} else {
					c.JSON(200, def)
				}
			})
		}

	}
	htmxGroup := engine.Group("/api/htmx")
	{
		htmxGroup.GET("/index", func(c *gin.Context) {

			c.HTML(200, "template/application.small", gin.H{
				"Application": nil,
				"Server":      nil,
				"Healthy":     rand.Intn(2) == 1,
			})
		})
	}
	engine.Run(":8080")
}
