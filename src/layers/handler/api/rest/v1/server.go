package v1

import (
	data "kukus/nam/v2/layers/data"
	handlers "kukus/nam/v2/layers/handler"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ServerController struct {
	Database *data.Database
}

func NewServerController(db *data.Database) *ServerController {
	return &ServerController{
		Database: db,
	}
}

// Initializes new Controller on declared RouterGroup, with specified resources
func (sc *ServerController) Init(routerGroup *gin.RouterGroup) {
	routerGroup.POST("/", sc.NewServer)
	routerGroup.GET("/", sc.GetAll)
	routerGroup.PATCH("/", handlers.MethodNotAllowed)
	routerGroup.PUT("/", handlers.MethodNotAllowed)
	routerGroup.DELETE("/", handlers.MethodNotAllowed)
	idGroup := routerGroup.Group("/:serverId")
	{
		idGroup.POST("/", handlers.MethodNotAllowed)
		idGroup.GET("/", sc.GetById)
		idGroup.PATCH("/", handlers.MethodNotImplemented)
		idGroup.PUT("/", sc.UpdateById)
		idGroup.DELETE("/", sc.RemoveById)
	}
}

// GetAll Servers
func (sc *ServerController) GetAll(ctx *gin.Context) {
	dtos, err := data.GetServerAll(sc.Database.Pool)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Unable to read server list", "trace": err.Error()})
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
	dtos, err := data.GetServerById(sc.Database.Pool, uint(serverId))
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Unable to read server list", "trace": err.Error()})
		return
	} else if dtos == nil {
		ctx.AbortWithStatus(404)
	} else {
		ctx.JSON(200, dtos)
	}
}

// Create Server
func (sc *ServerController) NewServer(ctx *gin.Context) {
	var server data.Server
	if err := ctx.ShouldBindJSON(&server); err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid JSON", "trace": err.Error()})
		return
	}
	id, err := server.DbInsert(sc.Database.Pool)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Unable to create Server", "trace": err.Error()})
		return
	}
	ctx.Header("HX-Redirect", "/servers")
	ctx.JSON(201, id)
}

// Get server with ID
func (sc *ServerController) RemoveById(ctx *gin.Context) {
	serverId, err := strconv.Atoi(ctx.Param("serverId"))
	if err != nil {
		ctx.JSON(400, gin.H{"error": "Must include ID of server"})
	}
	err = data.ServerDeleteById(sc.Database.Pool, uint(serverId))
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Unable to remove server", "trace": err.Error()})
		return
	} else {
		ctx.Header("HX-Redirect", "/servers")
		ctx.Status(200)
	}
}

// Get server with ID
func (sc *ServerController) UpdateById(ctx *gin.Context) {
	var server data.Server
	if err := ctx.ShouldBindJSON(&server); err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid JSON", "trace": err.Error()})
		return
	}
	err := server.Update(sc.Database.Pool)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Unable to update Server", "trace": err.Error()})
		return
	}
	ctx.Header("HX-Redirect", "/servers")
	ctx.JSON(200, server)
}
