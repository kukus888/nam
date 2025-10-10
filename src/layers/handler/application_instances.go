package handlers

import (
	"kukus/nam/v2/layers/data"
	"strconv"

	"github.com/gin-gonic/gin"
)

type InstanceView struct {
	Database *data.Database
}

func NewInstanceView(database *data.Database) InstanceView {
	return InstanceView{
		Database: database,
	}
}

// Renders the details page for a specific application instance
func (iv InstanceView) GetPageApplicationInstanceDetails(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil || id == 0 {
		ctx.AbortWithStatusJSON(500, gin.H{"error": "Invalid instance id", "trace": err.Error()})
		return
	}
	appInstance, err := data.GetApplicationInstanceFullById(iv.Database.Pool, uint64(id))
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": "", "trace": err.Error()})
		return
	} else if appInstance == nil {
		ctx.AbortWithStatusJSON(404, gin.H{"error": "Application instance not found"})
		return
	}

	// Get instance variables
	variables, err := data.GetApplicationInstanceVariablesByApplicationInstanceId(iv.Database.Pool, uint64(appInstance.Id))
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": "Unable to get application instance variables", "trace": err.Error()})
		return
	}

	// Get healthcheck template if available
	var healthcheck *data.Healthcheck
	if appInstance.ApplicationDefinition.HealthcheckId != nil {
		healthcheck, err = data.GetHealthCheckById(iv.Database.Pool, uint(*appInstance.ApplicationDefinition.HealthcheckId))
		if err != nil {
			ctx.AbortWithStatusJSON(500, gin.H{"error": "Unable to get healthcheck template", "trace": err.Error()})
			return
		}
	}

	// Get Healthcheck Results
	healthcheckResults, err := data.GetHealthcheckResultsByApplicationInstanceId(iv.Database.Pool, uint64(appInstance.Id))
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": "Unable to get healthcheck results", "trace": err.Error()})
		return
	}

	ctx.HTML(200, "pages/application/instance/details", gin.H{
		"Instance":           appInstance,
		"Variables":          *variables,
		"Healthcheck":        healthcheck,
		"HealthcheckResults": healthcheckResults,
	})
}

// GetPageApplicationInstanceVariables renders the page to view and edit variables for an application instance
func (iv InstanceView) GetPageApplicationInstanceVariables(ctx *gin.Context) {
	instanceId, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{"error": "Request must contain application instance id", "trace": err.Error()})
		return
	}
	instance, err := data.GetApplicationInstanceFullById(iv.Database.Pool, uint64(instanceId))
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": "Unable to get application instance", "trace": err.Error()})
		return
	}
	if instance == nil {
		ctx.AbortWithStatusJSON(404, gin.H{"error": "Application instance not found"})
		return
	}
	variables, err := data.GetApplicationInstanceVariablesByApplicationInstanceId(iv.Database.Pool, uint64(instance.Id))
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": "Unable to get application instance variables", "trace": err.Error()})
		return
	}
	ctx.HTML(200, "pages/application/instance/variables", gin.H{
		"Instance":  instance,
		"Variables": variables,
	})
}
