package v1

import (
	data "kukus/nam/v2/layers/data"
	"strconv"

	"github.com/gin-gonic/gin"
)

// RestApiAppDefVariablesController handles API requests related to application definition variables
type RestApiAppDefVariablesController struct {
	Database *data.Database
}

func NewAppDefVariablesController(db *data.Database) *RestApiAppDefVariablesController {
	return &RestApiAppDefVariablesController{
		Database: db,
	}
}

// GetAllVariables retrieves all variables for a specific application definition
func (ac *RestApiAppDefVariablesController) GetAllVariables(ctx *gin.Context) {
	appDefIdStr := ctx.Param("appId")
	if appDefIdStr == "" {
		ctx.JSON(400, gin.H{"error": "Application Definition ID is required"})
		return
	}
	appDefId, err := strconv.ParseUint(appDefIdStr, 10, 64)
	if err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid Application Definition ID", "trace": err.Error()})
		return
	}
	variables, err := data.GetApplicationDefinitionVariablesByApplicationDefinitionId(ac.Database.Pool, appDefId)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Unable to retrieve ApplicationDefinition variables", "trace": err.Error()})
		return
	}
	ctx.JSON(200, variables)
}

// CreateVariable creates new variable for a specific application definition
func (ac *RestApiAppDefVariablesController) CreateVariable(ctx *gin.Context) {
	var variable data.ApplicationDefinitionVariableDAO
	if err := ctx.ShouldBindJSON(&variable); err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid JSON", "trace": err.Error()})
		return
	}
	err := data.CreateApplicationDefinitionVariable(ac.Database.Pool, &variable)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Unable to create ApplicationDefinition variable", "trace": err.Error()})
		return
	}
	ctx.Status(201) // Created
}

// UpdateVariable updates a variable for a specific application definition
func (ac *RestApiAppDefVariablesController) UpdateVariable(ctx *gin.Context) {
	var variable data.ApplicationDefinitionVariableDAO
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
		ctx.JSON(400, gin.H{"error": "Variable ID in body does not match URL", "trace": err.Error()})
		return
	}
	err = data.UpdateApplicationDefinitionVariable(ac.Database.Pool, &variable)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Unable to update ApplicationDefinition variable", "trace": err.Error()})
		return
	}
	ctx.Status(204) // No Content
}

// DeleteVariable deletes a variable by its ID
func (ac *RestApiAppDefVariablesController) DeleteVariable(ctx *gin.Context) {
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
	err = data.DeleteApplicationDefinitionVariableById(ac.Database.Pool, varId)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Unable to delete ApplicationDefinition variable", "trace": err.Error()})
		return
	}
	ctx.Status(204) // No Content
}
