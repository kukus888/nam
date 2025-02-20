package v1

import (
	data "kukus/nam/v2/layers/data"
	handlers "kukus/nam/v2/layers/handler"
	services "kukus/nam/v2/layers/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ServerController struct {
	Service services.ServerService
}

func NewServerController(db *data.Database) *ServerController {
	return &ServerController{
		Service: services.ServerService{
			Database: db,
		},
	}
}

// Initializes new Controller on declared RouterGroup, with specified resources
func (ac *ServerController) Init(routerGroup *gin.RouterGroup) {
	routerGroup.POST("/", ac.NewServer)
	routerGroup.GET("/", ac.GetAll)
	routerGroup.PATCH("/", handlers.MethodNotAllowed)
	routerGroup.PUT("/", handlers.MethodNotAllowed)
	routerGroup.DELETE("/", handlers.MethodNotAllowed)
	idGroup := routerGroup.Group("/:serverId")
	{
		idGroup.POST("/", handlers.MethodNotAllowed)
		idGroup.GET("/", ac.GetById)
		idGroup.PATCH("/", handlers.MethodNotImplemented)
		idGroup.PUT("/", handlers.MethodNotImplemented)
		idGroup.DELETE("/", handlers.MethodNotImplemented)
		NewApplicationInstanceController(ac.Service.Database).Init(idGroup.Group("/instances"))
	}
}

// GetAll Servers
func (sc *ServerController) GetAll(ctx *gin.Context) {
	dtos, err := sc.Service.GetAllServers()
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Unable to read server list", "trace": err})
		return
	}
	ctx.JSON(200, dtos)
}

// Get server with ID
func (sc *ServerController) GetById(ctx *gin.Context) {
	serverId, err := strconv.Atoi(ctx.Param("serverId"))
	if err != nil {
		ctx.JSON(400, gin.H{"error": "Must include ID of server"})
	}
	dtos, err := sc.Service.GetServerById(uint(serverId))
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Unable to read server list", "trace": err})
		return
	} else if dtos == nil {
		ctx.AbortWithStatus(404)
	} else {
		ctx.JSON(200, dtos)
	}
}

// Create Server
func (sc *ServerController) NewServer(ctx *gin.Context) {
	var server data.ServerDAO
	if err := ctx.ShouldBindJSON(&server); err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid JSON", "trace": err})
		return
	}
	id, err := sc.Service.CreateServer(server)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Unable to create Server", "trace": err})
		return
	}
	ctx.JSON(201, id)
}
