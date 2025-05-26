package v1

import (
	data "kukus/nam/v2/layers/data"
	handlers "kukus/nam/v2/layers/handler"
	services "kukus/nam/v2/layers/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type HealthcheckController struct {
	Service services.HealthcheckService
}

func NewHealthcheckController(db *data.Database) *HealthcheckController {
	return &HealthcheckController{
		Service: services.HealthcheckService{
			Database: db,
		},
	}
}

// Initializes new Controller on declared RouterGroup, with specified resources
func (ac *HealthcheckController) Init(routerGroup *gin.RouterGroup) {
	routerGroup.POST("/", ac.NewHealthcheck)
	routerGroup.GET("/", ac.GetAll)
	routerGroup.PATCH("/", handlers.MethodNotAllowed)
	routerGroup.PUT("/", handlers.MethodNotAllowed)
	routerGroup.DELETE("/", handlers.MethodNotAllowed)
	idGroup := routerGroup.Group("/:hcId")
	{
		idGroup.POST("/", handlers.MethodNotAllowed)
		idGroup.GET("/", ac.GetById)
		idGroup.PATCH("/", handlers.MethodNotImplemented)
		idGroup.PUT("/", ac.UpdateHealthcheck)
		idGroup.DELETE("/", ac.Delete)
		NewApplicationInstanceController(ac.Service.Database).Init(idGroup.Group("/instances"))
	}
}

// GetAll Healthcheck
func (ac *HealthcheckController) GetAll(ctx *gin.Context) {
	dtos, err := data.GetHealthChecksAll(ac.Service.Database.Pool)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Unable to read application list", "trace": err})
		return
	}
	ctx.JSON(200, dtos)
}

// GetAll Healthcheck
func (ac *HealthcheckController) GetById(ctx *gin.Context) {
	hcId, err := strconv.Atoi(ctx.Param("hcId"))
	if err != nil {
		ctx.JSON(400, gin.H{"error": "Must include ID of application"})
	}
	dtos, err := data.GetHealthCheckById(ac.Service.Database.Pool, uint(hcId))
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Unable to read application list", "trace": err})
		return
	} else if dtos == nil {
		ctx.AbortWithStatus(404)
	} else {
		ctx.JSON(200, dtos)
	}
}

// Create Healthcheck
func (ac *HealthcheckController) NewHealthcheck(ctx *gin.Context) {
	var dto data.HealthcheckDTO
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid JSON", "trace": err})
		return
	}
	hc := dto.ToHealthcheck()
	id, err := hc.DbInsert(ac.Service.Database.Pool)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Unable to create Healthcheck", "trace": err})
		return
	}
	ctx.JSON(201, id)
}

// Update whole Healthcheck
func (ac *HealthcheckController) UpdateHealthcheck(ctx *gin.Context) {
	var dto data.HealthcheckDTO
	err := ctx.ShouldBindBodyWithJSON(&dto)
	if err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid JSON", "trace": err.Error()})
		return
	}
	dao := dto.ToHealthcheck()
	if err != nil {
		ctx.JSON(400, gin.H{"error": "Error converting DTO to DAO", "trace": err.Error()})
		return
	}
	err = dao.Update(ac.Service.Database.Pool)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Unable to update Healthcheck", "trace": err.Error()})
		return
	}
	ctx.JSON(200, dto)
}

// Delete Healthcheck
func (ac *HealthcheckController) Delete(ctx *gin.Context) {
	hcId, err := strconv.Atoi(ctx.Param("hcId"))
	if err != nil {
		ctx.JSON(400, gin.H{"error": "Must include ID of Healthcheck"})
		return
	}
	err = data.DeleteHealthCheckById(ac.Service.Database.Pool, uint(hcId))
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Unable to delete Healthcheck", "trace": err.Error()})
		return
	}
	ctx.JSON(204, gin.H{"message": "Healthcheck deleted successfully"})
}
