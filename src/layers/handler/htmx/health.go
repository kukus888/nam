package htmx

import (
	"kukus/nam/v2/layers/data"
	"slices"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// This file handles requests for healthcheck HTMX components.

type HtmxHealthHandler struct {
	Database *pgxpool.Pool
}

func NewHtmxHealthHandler(database *pgxpool.Pool) *HtmxHealthHandler {
	return &HtmxHealthHandler{
		Database: database,
	}
}

func (h *HtmxHealthHandler) Init(routeGroup *gin.RouterGroup) {
	allowedComponentSizes := []string{"tiny", "small", "medium", "large"}
	// Find out, which component is being requested, and return it filled with appropriate data.
	routeGroup.GET("/application/instance", func(ctx *gin.Context) {
		instanceIdStr := ctx.Query("id")
		liveReload := ctx.Query("live_reload") == "true" // Optional
		size := ctx.Query("size")

		// Parse and validate inputs
		instanceId, err := strconv.Atoi(instanceIdStr)
		if err != nil && instanceId != 0 {
			ctx.AbortWithStatusJSON(400, gin.H{"error": "Invalid instance id: " + err.Error()})
			return
		}
		if size != "" && !slices.Contains(allowedComponentSizes, size) {
			ctx.AbortWithStatusJSON(400, gin.H{"error": "Invalid size parameter. Allowed values are: " + strings.Join(allowedComponentSizes, ", ")})
			return
		}
		// Get health
		result, err := data.HealthcheckGetLatestResultByApplicationInstanceId(h.Database, uint64(instanceId))
		if err != nil {
			ctx.AbortWithStatusJSON(500, gin.H{"error": "Failed to get healthcheck result: " + err.Error()})
			return
		}
		// Get instance
		instance, err := data.GetApplicationInstanceFullById(h.Database, uint64(instanceId))
		if err != nil {
			ctx.AbortWithStatusJSON(500, gin.H{"error": "Failed to get application instance", "trace": err.Error()})
			return
		}
		// Render template with health data
		ctx.HTML(200, "components/health.application.instance."+size, gin.H{
			"Instance":     instance, // This should be replaced with actual instance data
			"Result":       result,
			"LiveReload":   liveReload,
			"Healthy":      result.IsSuccessful,
			"ResponseTime": result.ResTime,
			"Timestamp":    result.TimeEnd,
		})
	})
	routeGroup.GET("/application/definition", func(ctx *gin.Context) {
		definitionIdStr := ctx.Query("id")
		liveReload := ctx.Query("live_reload") == "true" // Optional
		size := ctx.Query("size")

		// Parse and validate inputs
		definitionId, err := strconv.Atoi(definitionIdStr)
		if err != nil && definitionId != 0 {
			ctx.AbortWithStatusJSON(400, gin.H{"error": "Invalid definition id: " + err.Error()})
			return
		}
		if size != "" && !slices.Contains(allowedComponentSizes, size) {
			ctx.AbortWithStatusJSON(400, gin.H{"error": "Invalid size parameter. Allowed values are: " + strings.Join(allowedComponentSizes, ", ")})
			return
		}
		// Get health
		results, err := data.HealthcheckGetLatestResultByApplicationDefinitionId(h.Database, uint64(definitionId))
		if err != nil {
			ctx.AbortWithStatusJSON(500, gin.H{"error": "Failed to get healthcheck results: " + err.Error()})
			return
		}
		// Compute stats
		healthyCount := 0
		for _, result := range *results {
			if result.IsSuccessful {
				healthyCount++
			}
		}
		ctx.HTML(200, "components/health.application.definition."+size, gin.H{
			"Id":           definitionId,
			"LiveReload":   liveReload,
			"HealthyCount": healthyCount,
			"TotalCount":   len(*results),
		})
	})
	routeGroup.GET("/application/definition_with_instances", func(ctx *gin.Context) {
		definitionIdStr := ctx.Query("id")
		liveReload := ctx.Query("live_reload") == "true" // Optional
		size := ctx.Query("size")

		// Parse and validate inputs
		definitionId, err := strconv.Atoi(definitionIdStr)
		if err != nil && definitionId != 0 {
			ctx.AbortWithStatusJSON(400, gin.H{"error": "Invalid definition id", "trace": err.Error()})
			return
		}
		if size != "" && !slices.Contains(allowedComponentSizes, size) {
			ctx.AbortWithStatusJSON(400, gin.H{"error": "Invalid size parameter. Allowed values are: " + strings.Join(allowedComponentSizes, ", ")})
			return
		}
		// Get definition
		definition, err := data.GetApplicationDefinitionById(h.Database, uint64(definitionId))
		if err != nil {
			ctx.AbortWithStatusJSON(500, gin.H{"error": "Failed to get application definition", "trace": err.Error()})
			return
		}
		// Get instances
		instances, err := data.GetApplicationInstancesByApplicationDefinitionId(h.Database, uint64(definitionId))
		if err != nil {
			ctx.AbortWithStatusJSON(500, gin.H{"error": "Failed to get application instances", "trace": err.Error()})
			return
		}
		// Get health
		results, err := data.HealthcheckGetLatestResultByApplicationDefinitionId(h.Database, uint64(definitionId))
		if err != nil {
			ctx.AbortWithStatusJSON(500, gin.H{"error": "Failed to get healthcheck results", "trace": err.Error()})
			return
		}
		// Compute stats
		healthyCount := 0
		for _, result := range *results {
			if result.IsSuccessful {
				healthyCount++
			}
		}
		ctx.HTML(200, "components/health.application.definition.withInstances."+size, gin.H{
			"Id":           definitionId,
			"Definition":   definition,
			"Instances":    instances,
			"LiveReload":   liveReload,
			"HealthyCount": healthyCount,
			"TotalCount":   len(*results),
		})
	})
}
