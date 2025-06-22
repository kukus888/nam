package v1

import (
	data "kukus/nam/v2/layers/data"
	handlers "kukus/nam/v2/layers/handler"
	"strconv"

	"github.com/gin-gonic/gin"
)

type RestApiApplicationController struct {
	Database *data.Database
}

func NewApplicationController(db *data.Database) *RestApiApplicationController {
	return &RestApiApplicationController{
		Database: db,
	}
}

// Initializes new Controller on declared RouterGroup, with specified resources
func (ac *RestApiApplicationController) Init(routerGroup *gin.RouterGroup) {
	routerGroup.POST("/", ac.NewApplication)
	routerGroup.GET("/", ac.GetAll)
	routerGroup.PATCH("/", handlers.MethodNotAllowed)
	routerGroup.PUT("/", handlers.MethodNotAllowed)
	routerGroup.DELETE("/", handlers.MethodNotAllowed)
	idGroup := routerGroup.Group("/:appId")
	{
		idGroup.POST("/", handlers.MethodNotAllowed)
		idGroup.GET("/", ac.GetById)
		idGroup.PATCH("/", handlers.MethodNotImplemented)
		idGroup.PUT("/", ac.UpdateApplicationDefinition)
		idGroup.DELETE("/", ac.DeleteById)
		NewApplicationInstanceController(ac.Database).Init(idGroup.Group("/instances"))
	}
}

// GetAll ApplicationDefinition
func (ac *RestApiApplicationController) GetAll(ctx *gin.Context) {
	dtos, err := data.GetApplicationDefinitionsAll(ac.Database.Pool)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Unable to read application list", "trace": err.Error()})
		return
	}
	ctx.JSON(200, dtos)
}

func (ac *RestApiApplicationController) DeleteById(ctx *gin.Context) {
	appId, err := strconv.Atoi(ctx.Param("appId"))
	if err != nil {
		ctx.JSON(400, gin.H{"error": "Must include ID of application"})
		return
	}
	err = data.DeleteApplicationDefinitionById(ac.Database.Pool, uint64(appId))
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Unable to delete application", "trace": err.Error()})
		return
	}
	ctx.Status(204) // No Content
}

// GetAll ApplicationDefinition
func (ac *RestApiApplicationController) GetById(ctx *gin.Context) {
	appId, err := strconv.Atoi(ctx.Param("appId"))
	if err != nil {
		ctx.JSON(400, gin.H{"error": "Must include ID of application"})
	}
	dtos, err := data.GetApplicationDefinitionById(ac.Database.Pool, uint64(appId))
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Unable to read application list", "trace": err.Error()})
		return
	} else if dtos == nil {
		ctx.AbortWithStatus(404)
	} else {
		ctx.JSON(200, dtos)
	}
}

// Create ApplicationDefinition
func (ac *RestApiApplicationController) NewApplication(ctx *gin.Context) {
	var appDef data.ApplicationDefinitionDAO
	if err := ctx.ShouldBindJSON(&appDef); err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid JSON", "trace": err.Error()})
		return
	}
	id, err := appDef.DbInsert(ac.Database.Pool)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Unable to create ApplicationDefinition", "trace": err.Error()})
		return
	}
	ctx.JSON(201, id)
}

// Update ApplicationDefinition
func (ac *RestApiApplicationController) UpdateApplicationDefinition(ctx *gin.Context) {
	var appDef data.ApplicationDefinitionDAO
	if err := ctx.ShouldBindJSON(&appDef); err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid JSON", "trace": err.Error()})
		return
	}
	err := data.UpdateApplicationDefinition(ac.Database.Pool, &appDef)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Unable to update ApplicationDefinition", "trace": err.Error()})
		return
	}
	ctx.Status(204) // No Content
}
