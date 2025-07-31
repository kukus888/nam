package handlers

import (
	"kukus/nam/v2/layers/data"
	"strconv"

	"github.com/gin-gonic/gin"
)

type HealthcheckView struct {
	Database *data.Database
}

func NewHealthcheckView(database *data.Database) HealthcheckView {
	return HealthcheckView{
		Database: database,
	}
}

// GetPageHealthchecks renders the page for listing all health checks.
func (av HealthcheckView) GetPageHealthchecks(ctx *gin.Context) {
	hcs, err := data.GetHealthChecksAll(av.Database.Pool)
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}
	ctx.HTML(200, "pages/healthchecks", gin.H{
		"Healthchecks": hcs,
	})
}

// GetPageHealthcheckCreate renders the page for creating a new health check.
func (av HealthcheckView) GetPageHealthcheckCreate(ctx *gin.Context) {
	hcs, err := data.GetHealthChecksAll(av.Database.Pool)
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}
	ctx.HTML(200, "pages/healthchecks/create", gin.H{
		"Healthchecks": &hcs,
	})
}

// GetPageHealthcheckDetails renders the details page for a specific health check.
func (av HealthcheckView) GetPageHealthcheckDetails(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}
	hc, err := data.GetHealthCheckById(av.Database.Pool, uint(id))
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}
	ctx.HTML(200, "pages/healthchecks/details", gin.H{
		"Healthcheck": hc,
	})
}

// GetPageHealthcheckEdit renders the page for editing an existing health check.
func (av HealthcheckView) GetPageHealthcheckEdit(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}
	hc, err := data.GetHealthCheckById(av.Database.Pool, uint(id))
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}
	ctx.HTML(200, "pages/healthchecks/edit", gin.H{
		"Healthcheck": hc,
	})
}
