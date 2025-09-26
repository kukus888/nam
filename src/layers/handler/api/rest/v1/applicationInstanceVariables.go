package v1

import (
	data "kukus/nam/v2/layers/data"
	"strconv"

	"github.com/gin-gonic/gin"
)

// RestApiAppInstanceVariablesController handles API requests related to application instance variables
type RestApiAppInstanceVariablesController struct {
	Database *data.Database
}

func NewAppInstanceVariablesController(db *data.Database) *RestApiAppInstanceVariablesController {
	return &RestApiAppInstanceVariablesController{
		Database: db,
	}
}

// GetAllVariables retrieves all variables for a specific application instance
func (ac *RestApiAppInstanceVariablesController) GetAllVariables(ctx *gin.Context) {
	appInstanceIdStr := ctx.Param("instanceId")
	if appInstanceIdStr == "" {
		ctx.JSON(400, gin.H{"error": "Application Instance ID is required"})
		return
	}
	appInstanceId, err := strconv.ParseUint(appInstanceIdStr, 10, 64)
	if err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid Application Instance ID", "trace": err.Error()})
		return
	}
	variables, err := data.GetApplicationInstanceVariablesByApplicationInstanceId(ac.Database.Pool, appInstanceId)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Unable to retrieve ApplicationInstance variables", "trace": err.Error()})
		return
	}
	ctx.JSON(200, variables)
}

// CreateVariable creates new variable for a specific application instance
func (ac *RestApiAppInstanceVariablesController) CreateVariable(ctx *gin.Context) {
	var variable data.ApplicationInstanceVariableDAO
	if err := ctx.ShouldBindJSON(&variable); err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid JSON", "trace": err.Error()})
		return
	}
	err := data.CreateApplicationInstanceVariable(ac.Database.Pool, &variable)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Unable to create ApplicationInstance variable", "trace": err.Error()})
		return
	}
	ctx.Status(201) // Created
}

// UpdateVariable updates a variable for a specific application instance
func (ac *RestApiAppInstanceVariablesController) UpdateVariable(ctx *gin.Context) {
	var variable data.ApplicationInstanceVariableDAO
	if err := ctx.ShouldBindJSON(&variable); err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid JSON", "trace": err.Error()})
		return
	}
	// Check if the ID in body is the same as in the URL
	varIdStr := ctx.Param("varId")
	varId, err := strconv.ParseUint(varIdStr, 10, 64)
	if err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid Variable ID in URL", "trace": err.Error()})
		return
	}
	if variable.Id != uint(varId) {
		ctx.JSON(400, gin.H{"error": "Variable ID in body does not match URL"})
		return
	}
	err = data.UpdateApplicationInstanceVariable(ac.Database.Pool, &variable)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Unable to update ApplicationInstance variable", "trace": err.Error()})
		return
	}
	ctx.Status(204) // No Content
}

// DeleteVariable deletes a variable by its ID
func (ac *RestApiAppInstanceVariablesController) DeleteVariable(ctx *gin.Context) {
	varIdStr := ctx.Param("varId")
	if varIdStr == "" {
		ctx.JSON(400, gin.H{"error": "Variable ID is required"})
		return
	}
	varId, err := strconv.ParseUint(varIdStr, 10, 64)
	if err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid Variable ID", "trace": err.Error()})
		return
	}
	err = data.DeleteApplicationInstanceVariableById(ac.Database.Pool, varId)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Unable to delete ApplicationInstance variable", "trace": err.Error()})
		return
	}
	ctx.Status(204) // No Content
}
