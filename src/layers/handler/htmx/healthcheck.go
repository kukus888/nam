package htmx

import (
	"kukus/nam/v2/layers/data"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type HealthcheckView struct {
	Database *data.Database
}

func NewHealthcheckView(db *data.Database) HealthcheckView {
	return HealthcheckView{
		Database: db,
	}
}

func (hcv HealthcheckView) Init(routeGroup *gin.RouterGroup) {
	idGroup := routeGroup.Group("/:id")
	{
		idGroup.GET("/tiny", hcv.GetHealthCheckTiny)
	}
	routeGroup.GET("/records/latest", hcv.GetHealthCheckLatest)
	routeGroup.GET("/records/:id", hcv.GetHealthCheckLatest)
}

func (hcv HealthcheckView) GetHealthCheckTiny(ctx *gin.Context) {
	hcId, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}
	hc := data.HealthcheckRecord{
		ID:               uint64(hcId),
		Timestamp:        time.Now(),
		HttpResponseCode: 200,
		HttpResponseBody: "OK",
		Status:           "healthy",
	}
	ctx.HTML(200, "template/healthcheck.tiny", gin.H{
		"HealthcheckRecord": hc,
	})
}

// GetHealthCheckLatest returns the latest health check record for a given instance id
// Size agnostic, will return the health check record in the requested size
func (hcv HealthcheckView) GetHealthCheckLatest(ctx *gin.Context) {
	instanceId, err := strconv.ParseUint(ctx.Query("instance_id"), 10, 64)
	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}
	size := ctx.Query("size")
	if size == "" {
		ctx.AbortWithStatusJSON(400, gin.H{"error": "size is required"})
		return
	}
	// TODO: get latest health, actual data
	hc := data.HealthcheckRecord{
		ID:               instanceId,
		Timestamp:        time.Now(),
		HttpResponseCode: 200,
		HttpResponseBody: "OK",
		Status:           "healthy",
	}
	ctx.HTML(200, "template/healthcheck.latest."+size, gin.H{
		"HealthcheckRecord": hc,
		"InstanceId":        instanceId,
	})
}

// GetHealthCheckById returns the health check record for a given healthcheck record id
// Size agnostic, will return the health check record in the requested size
func (hcv HealthcheckView) GetHealthCheckById(ctx *gin.Context) {
	instanceId, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}
	size := ctx.Query("size")
	if size == "" {
		ctx.AbortWithStatusJSON(400, gin.H{"error": "size is required"})
		return
	}
	// TODO: get latest health, actual data
	hc := data.HealthcheckRecord{
		ID:               instanceId,
		Timestamp:        time.Now(),
		HttpResponseCode: 200,
		HttpResponseBody: "OK",
		Status:           "healthy",
	}
	ctx.HTML(200, "template/healthcheck.latest."+size, gin.H{
		"HealthcheckRecord": hc,
		"InstanceId":        instanceId,
	})
}
