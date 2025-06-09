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

/*
 *	Component used for viewing Application instance pages
 */
func (iv InstanceView) Init(routeGroup *gin.RouterGroup) {
	routeGroup.GET("/:id/details", iv.GetPageApplicationInstanceDetails)
}

// Returns the details page for a specific application instance
func (iv InstanceView) GetPageApplicationInstanceDetails(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil || id == 0 {
		ctx.AbortWithStatusJSON(500, gin.H{"error": "Invalid instance id", "trace": err.Error()})
		return
	}
	appInstance, err := data.GetApplicationInstanceById(iv.Database.Pool, uint64(id))
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": "", "trace": err.Error()})
		return
	} else if appInstance == nil {
		ctx.AbortWithStatusJSON(404, gin.H{"error": "Application instance not found"})
		return
	}
	healthcheckResults, err := data.GetHealthcheckResultsByApplicationInstanceId(iv.Database.Pool, uint64(appInstance.Id))
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": "Unable to get healthcheck results.", "trace": err.Error()})
		return
	}
	ctx.HTML(200, "pages/application/instance/details", gin.H{
		"Instance":           appInstance,
		"HealthcheckResults": *healthcheckResults,
	})
}
