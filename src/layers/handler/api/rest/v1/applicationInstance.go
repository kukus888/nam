package v1

import (
	data "kukus/nam/v2/layers/data"
	handlers "kukus/nam/v2/layers/handler"
	services "kukus/nam/v2/layers/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ApplicationInstanceController struct {
	Service services.ApplicationInstanceService
}

func NewApplicationInstanceController(db *data.Database) *ApplicationInstanceController {
	return &ApplicationInstanceController{
		Service: services.ApplicationInstanceService{
			Database: db,
		},
	}
}

// Initializes new Controller on declared RouterGroup, with specified resources
func (aic *ApplicationInstanceController) Init(routerGroup *gin.RouterGroup) {
	routerGroup.POST("/", aic.CreateInstance)
	routerGroup.GET("/", aic.GetAllInstances)
	routerGroup.PATCH("/", handlers.MethodNotAllowed)
	routerGroup.PUT("/", handlers.MethodNotAllowed)
	routerGroup.DELETE("/", handlers.MethodNotAllowed)
	idGroup := routerGroup.Group("/:instanceId")
	{
		idGroup.POST("/", handlers.MethodNotAllowed)
		idGroup.GET("/", aic.GetById)
		idGroup.PATCH("/", handlers.MethodNotImplemented)
		idGroup.PUT("/", handlers.MethodNotImplemented)
		idGroup.DELETE("/", aic.DeleteInstance)
	}
}

// Creates new ApplicationInstance
func (aic *ApplicationInstanceController) CreateInstance(ctx *gin.Context) {
	var appInst data.ApplicationInstanceDAO
	if err := ctx.ShouldBindJSON(&appInst); err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid JSON", "trace": err})
		return
	}
	dtos, err := aic.Service.CreateApplicationInstance(appInst)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Unable to read application list", "trace": err})
		return
	} else if dtos == nil {
		ctx.AbortWithStatus(404)
	} else {
		ctx.JSON(200, dtos)
	}
}

// GetAll ApplicationInstance
func (aic *ApplicationInstanceController) GetById(ctx *gin.Context) {
	instanceId, err := strconv.Atoi(ctx.Param("instanceId"))
	if err != nil {
		ctx.JSON(400, gin.H{"error": "Must include ID of application instance"})
	}
	dtos, err := aic.Service.GetApplicationInstanceById(instanceId)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Unable to read application list", "trace": err})
		return
	} else if dtos == nil {
		ctx.AbortWithStatus(404)
	} else {
		ctx.JSON(200, dtos)
	}
}

// GetAll ApplicationInstance
func (aic *ApplicationInstanceController) GetAllInstances(ctx *gin.Context) {
	appId, err := strconv.Atoi(ctx.Param("appId"))
	dtos, err := aic.Service.GetAllApplicationInstances(appId)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Unable to read application list", "trace": err})
		return
	}
	ctx.JSON(200, dtos)
}

// Delete ApplicationInstance
func (aic *ApplicationInstanceController) DeleteInstance(ctx *gin.Context) {
	instanceId, err := strconv.Atoi(ctx.Param("instanceId"))
	if err != nil {
		ctx.JSON(400, gin.H{"error": "Must include ID of application instance"})
	}
	err = aic.Service.RemoveApplicationInstanceById(instanceId)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Unable to read application list", "trace": err})
		return
	}
	ctx.Status(200)
}
