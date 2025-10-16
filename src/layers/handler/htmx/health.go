package htmx

import (
	"kukus/nam/v2/layers/data"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// This file handles requests for healthcheck HTMX components.

type HtmxHealthHandler struct {
	Database              *pgxpool.Pool
	AllowedComponentSizes []string
}

func NewHtmxHealthHandler(database *pgxpool.Pool) *HtmxHealthHandler {
	return &HtmxHealthHandler{
		Database:              database,
		AllowedComponentSizes: []string{"tiny", "small", "medium", "large"},
	}
}

func (h *HtmxHealthHandler) Init(routeGroup *gin.RouterGroup) {
	routeGroup.GET("/application/instance", h.RenderHealthApplicationInstanceComponent)
	routeGroup.GET("/application/definition", h.RenderHealthApplicationDefinitionComponent)
	routeGroup.GET("/application/definition_with_instances", h.RenderHealthApplicationDefinitionWithInstancesComponent)
	routeGroup.GET("/healthcheck/result", h.RenderHealthCheckResultComponent)
	routeGroup.GET("/timeline", h.RenderHealthTimelineComponent)
}

// RenderHealthApplicationInstanceComponent renders the health component for a specific application definition.
// It expects the definition ID as a query parameter and a size parameter for the component size. Optional live reload can also be specified.
func (h *HtmxHealthHandler) RenderHealthApplicationDefinitionComponent(ctx *gin.Context) {
	definitionIdStr := ctx.Query("id")
	liveReload := ctx.Query("live_reload") == "true" // Optional
	size := ctx.Query("size")

	// Parse and validate inputs
	definitionId, err := strconv.Atoi(definitionIdStr)
	if err != nil && definitionId != 0 {
		ctx.AbortWithStatusJSON(400, gin.H{"error": "Invalid definition id: " + err.Error()})
		return
	}
	if size != "" && !slices.Contains(h.AllowedComponentSizes, size) {
		ctx.AbortWithStatusJSON(400, gin.H{"error": "Invalid size parameter. Allowed values are: " + strings.Join(h.AllowedComponentSizes, ", ")})
		return
	}
	// Get health
	results, err := data.GetHealthcheckLatestResultByApplicationDefinitionId(h.Database, uint64(definitionId))
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": "Failed to get healthcheck results: " + err.Error()})
		return
	}
	if results == nil {
		results = &[]data.ApplicationDefinitionHealthcheckResult{} // Ensure results is not nil
	}
	// Compute stats
	healthyCount := 0
	maintenanceCount := 0
	for _, result := range *results {
		if result.IsMaintenance {
			maintenanceCount++
			continue
		}
		if result.IsSuccessful {
			healthyCount++
		}
	}
	ctx.HTML(200, "components/health.application.definition."+size, gin.H{
		"Id":               definitionId,
		"LiveReload":       liveReload,
		"HealthyCount":     healthyCount,
		"MaintenanceCount": maintenanceCount,
		"TotalCount":       len(*results) - maintenanceCount,
	})
}

// RenderHealthApplicationInstanceComponent renders the health component for a specific application instance.
// It expects the instance ID as a query parameter and a size parameter for the component size. Optional live reload can also be specified.
func (h *HtmxHealthHandler) RenderHealthApplicationInstanceComponent(ctx *gin.Context) {
	instanceIdStr := ctx.Query("id")
	liveReload := ctx.Query("live_reload") == "true" // Optional
	size := ctx.Query("size")

	// Parse and validate inputs
	instanceId, err := strconv.Atoi(instanceIdStr)
	if err != nil && instanceId != 0 {
		ctx.AbortWithStatusJSON(400, gin.H{"error": "Invalid instance id: " + err.Error()})
		return
	}
	if size != "" && !slices.Contains(h.AllowedComponentSizes, size) {
		ctx.AbortWithStatusJSON(400, gin.H{"error": "Invalid size parameter. Allowed values are: " + strings.Join(h.AllowedComponentSizes, ", ")})
		return
	}
	// Get health
	result, err := data.HealthcheckGetLatestResultByApplicationInstanceId(h.Database, uint64(instanceId))
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": "Failed to get healthcheck result: " + err.Error()})
		return
	}
	if result == nil {
		result = &data.HealthcheckResult{} // Ensure result is not nil
	}
	// Get instance
	instance, err := data.GetApplicationInstanceFullById(h.Database, uint64(instanceId))
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": "Failed to get application instance", "trace": err.Error()})
		return
	}
	// Get healthcheckTemplate definition
	healthcheckTemplate, err := data.GetHealthCheckById(h.Database, result.HealthcheckID)
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": "Failed to get healthcheck definition", "trace": err.Error()})
		return
	}
	// Figure out which icon to use
	// TODO: Get from DB
	iconPath := "/static/icons/golang.svg"
	switch instance.Type {
	case "JBoss":
		iconPath = "/static/icons/jboss.svg"
	case "Springboot":
		iconPath = "/static/icons/spring.svg"
	}
	// Render template with health data
	ctx.HTML(200, "components/health.application.instance."+size, gin.H{
		"Instance":            instance,
		"HealthcheckTemplate": healthcheckTemplate,
		"Result":              result,
		"LiveReload":          liveReload,
		"Healthy":             result.IsSuccessful,
		"ResponseTime":        result.ResTime,
		"Timestamp":           result.TimeEnd,
		"IconPath":            iconPath,
	})
}

// RenderHealthApplicationDefinitionWithInstancesComponent renders the health component for a specific application definition with its instances.
// It expects the definition ID as a query parameter and a size parameter for the component size. Optional live reload can also be specified.
func (h *HtmxHealthHandler) RenderHealthApplicationDefinitionWithInstancesComponent(ctx *gin.Context) {
	definitionIdStr := ctx.Query("id")
	liveReload := ctx.Query("live_reload") == "true" // Optional
	size := ctx.Query("size")

	// Parse and validate inputs
	definitionId, err := strconv.Atoi(definitionIdStr)
	if err != nil && definitionId != 0 {
		ctx.AbortWithStatusJSON(400, gin.H{"error": "Invalid definition id", "trace": err.Error()})
		return
	}
	if size != "" && !slices.Contains(h.AllowedComponentSizes, size) {
		ctx.AbortWithStatusJSON(400, gin.H{"error": "Invalid size parameter. Allowed values are: " + strings.Join(h.AllowedComponentSizes, ", ")})
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
	results, err := data.GetHealthcheckLatestResultByApplicationDefinitionId(h.Database, uint64(definitionId))
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
}

func (h *HtmxHealthHandler) RenderHealthCheckResultComponent(ctx *gin.Context) {
	resultIdStr := ctx.Query("id")
	size := ctx.Query("size")

	// Parse and validate inputs
	resultId, err := strconv.Atoi(resultIdStr)
	if err != nil && resultId != 0 {
		ctx.AbortWithStatusJSON(400, gin.H{"error": "Invalid healthcheck result id", "trace": err.Error()})
		return
	}
	if size != "" && !slices.Contains(h.AllowedComponentSizes, size) {
		ctx.AbortWithStatusJSON(400, gin.H{"error": "Invalid size parameter. Allowed values are: " + strings.Join(h.AllowedComponentSizes, ", ")})
		return
	}
	result, err := data.GetHealthcheckResultById(h.Database, uint(resultId))
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": "Failed to get healthcheck result", "trace": err.Error()})
		return
	}
	ctx.HTML(200, "components/health.result."+size, gin.H{
		"Id":           resultId,
		"Result":       result,
		"IsSuccessful": result.IsSuccessful,
	})
}

// RenderHealthTimelineComponent renders the health timeline component for a specific application instance.
// It expects the instance ID as a query parameter and an optional time range parameter.
func (h *HtmxHealthHandler) RenderHealthTimelineComponent(ctx *gin.Context) {
	instanceIdStr := ctx.Query("instance_id")
	timeRange := ctx.Query("range") // Optional, e.g., "1h", "24h", "7d"

	// Parse and validate inputs
	instanceId, err := strconv.Atoi(instanceIdStr)
	if err != nil || instanceId == 0 {
		ctx.AbortWithStatusJSON(400, gin.H{"error": "Invalid instance id: " + err.Error()})
		return
	}

	// Get instance details
	instance, err := data.GetApplicationInstanceFullById(h.Database, uint64(instanceId))
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": "Failed to get application instance", "trace": err.Error()})
		return
	}
	// Compute time range filter
	var startTime time.Time
	var endTime time.Time
	if timeRange != "" {
		var duration time.Duration
		switch {
		case strings.HasSuffix(timeRange, "h"):
			hoursStr := strings.TrimSuffix(timeRange, "h")
			hours, err := strconv.Atoi(hoursStr)
			if err != nil {
				ctx.AbortWithStatusJSON(400, gin.H{"error": "Invalid time range format", "trace": err.Error()})
				return
			}
			duration = time.Duration(hours) * time.Hour
		case strings.HasSuffix(timeRange, "d"):
			daysStr := strings.TrimSuffix(timeRange, "d")
			days, err := strconv.Atoi(daysStr)
			if err != nil {
				ctx.AbortWithStatusJSON(400, gin.H{"error": "Invalid time range format", "trace": err.Error()})
				return
			}
			duration = time.Duration(days*24) * time.Hour
		default:
			ctx.AbortWithStatusJSON(400, gin.H{"error": "Invalid time range format. Use 'h' for hours or 'd' for days."})
			return
		}
		// Calculate the start time for filtering
		startTime = time.Now().Add(-duration)
		endTime = time.Now()
	} else {
		// For when there is start and end time range, TBD
		startTime = time.Now().Add(-24 * time.Hour)
		endTime = time.Now()
	}

	// Get healthcheck results for the instance
	healthcheckResults, err := data.GetHealthcheckResultsByApplicationInstanceIdRange(h.Database, uint64(instanceId), startTime, endTime)
	if err != nil || healthcheckResults == nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": "Failed to get healthcheck results", "trace": err.Error()})
		return
	}

	// Transform the data into buckets
	const bucketCount = 25
	const indicatorCount = 7 // Number of indicators to show on the timeline
	indicatorInterval := bucketCount / indicatorCount
	// Initialize timeline structure
	timeline := Timeline{
		Buckets:         make([]TimelineBucket, bucketCount),
		MaxResponseTime: 0,
	}
	bucketDuration := endTime.Sub(startTime) / bucketCount
	bucketSize := len(*healthcheckResults) / bucketCount
	if bucketSize == 0 {
		bucketSize = 1 // Ensure at least one item per bucket
	}
	for bucket := range bucketCount {
		// Fill this bucket with a 1/bucketCount slice of the time range
		// Slice input data
		sliceToProcess := (*healthcheckResults)[bucket*bucketSize : (bucket+1)*bucketSize]
		timeline.Buckets[bucket] = TimelineBucket{
			StartTime:   startTime.Add(bucketDuration * time.Duration(bucket)),
			EndTime:     startTime.Add(bucketDuration * time.Duration(bucket+1)),
			TotalChecks: 0,
			Successes:   0,
			Failures:    0,
		}
		for _, result := range sliceToProcess {
			timeline.Buckets[bucket].TotalChecks++
			if result.IsSuccessful {
				timeline.Buckets[bucket].Successes++
			} else {
				timeline.Buckets[bucket].Failures++
			}
			// Update max response time if needed
			if float64(result.ResTime) > timeline.MaxResponseTime {
				timeline.MaxResponseTime = float64(result.ResTime)
			}
			// Mark indicator on each 7th bucket
			if bucket%indicatorInterval == 0 {
				timeline.Buckets[bucket].Indicator = true
			}
		}
		// Compute average response time for the bucket
		var totalResTime float64
		for _, result := range sliceToProcess {
			totalResTime += float64(result.ResTime)
		}
		timeline.Buckets[bucket].AverageResTime = totalResTime / float64(timeline.Buckets[bucket].TotalChecks)
	}
	// Compute average response time percentage for each bucket
	for i := range timeline.Buckets {
		timeline.Buckets[i].AvgResTimePercent = int((timeline.Buckets[i].AverageResTime / timeline.MaxResponseTime) * 100)
		if timeline.Buckets[i].AvgResTimePercent < 10 {
			timeline.Buckets[i].AvgResTimePercent = 10 // Ensure minimum height for visibility
		}
	}

	// Render the timeline component
	ctx.HTML(200, "components/health_results_timeline", gin.H{
		"Instance":           instance,
		"HealthcheckResults": *healthcheckResults,
		"TimeRange":          timeRange,
		"Timeline":           timeline,
	})
}

type TimelineBucket struct {
	StartTime         time.Time
	EndTime           time.Time
	AverageResTime    float64
	AvgResTimePercent int // Percentage relative to the max response time in the timeline
	TotalChecks       int
	Successes         int
	Failures          int
	Indicator         bool // True if there is supposed to be a line indicator in the timeline
}

type Timeline struct {
	Buckets         []TimelineBucket
	MaxResponseTime float64
}
