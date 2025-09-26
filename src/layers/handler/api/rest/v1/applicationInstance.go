package v1

import (
	data "kukus/nam/v2/layers/data"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ApplicationInstanceController struct {
	DatabasePool *pgxpool.Pool
}

func NewApplicationInstanceController(db *data.Database) *ApplicationInstanceController {
	return &ApplicationInstanceController{
		DatabasePool: db.Pool,
	}
}

// Creates new ApplicationInstance
func (aic *ApplicationInstanceController) CreateInstance(ctx *gin.Context) {
	var appInst data.ApplicationInstance
	if err := ctx.ShouldBindJSON(&appInst); err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid JSON", "trace": err.Error()})
		return
	}
	dtos, err := data.CreateApplicationInstance(aic.DatabasePool, appInst)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Unable to create application instance", "trace": err.Error()})
		return
	} else if dtos == nil {
		ctx.AbortWithStatus(404)
	} else {
		ctx.Header("HX-Redirect", "/applications/"+strconv.Itoa(int(appInst.ApplicationDefinitionID))+"/details")
		ctx.JSON(200, dtos)
	}
}

// Get ApplicationInstance with Id
func (aic *ApplicationInstanceController) GetById(ctx *gin.Context) {
	instanceId, err := strconv.ParseUint(ctx.Param("instanceId"), 10, 64)
	if err != nil {
		ctx.JSON(400, gin.H{"error": "Must include ID of application instance"})
	}
	dtos, err := data.GetApplicationInstanceFullById(aic.DatabasePool, instanceId)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Unable to read application list", "trace": err.Error()})
		return
	} else if dtos == nil {
		ctx.AbortWithStatus(404)
	} else {
		ctx.JSON(200, dtos)
	}
}

// GetAll ApplicationInstance
func (aic *ApplicationInstanceController) GetAllInstances(ctx *gin.Context) {
	dtos, err := data.GetAllApplicationInstancesFull(aic.DatabasePool)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Unable to read application list", "trace": err.Error()})
		return
	}
	ctx.JSON(200, dtos)
}

// Delete ApplicationInstance
func (aic *ApplicationInstanceController) DeleteInstance(ctx *gin.Context) {
	instanceId, err := strconv.ParseUint(ctx.Param("instanceId"), 10, 64)
	if err != nil {
		ctx.JSON(400, gin.H{"error": "Must include ID of application instance"})
	}
	err = data.DeleteApplicationInstanceById(aic.DatabasePool, instanceId)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Unable to delete application instance", "trace": err.Error()})
		return
	}
	ctx.Status(200)
}

// Toggle maintenance mode
func (aic *ApplicationInstanceController) ToggleMaintenance(ctx *gin.Context) {
	var req data.ApplicationInstance
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid JSON", "trace": err.Error()})
		return
	}
	err := data.ToggleApplicationInstanceMaintenance(aic.DatabasePool, uint64(req.Id), req.MaintenanceMode)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Unable to toggle maintenance mode", "trace": err.Error()})
		return
	}
	ctx.Status(204)
}
