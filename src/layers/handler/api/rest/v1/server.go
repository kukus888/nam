package v1

import (
	data "kukus/nam/v2/layers/data"
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

// Update server by ID
func (sc *ServerController) UpdateById(ctx *gin.Context) {
	serverId, err := strconv.Atoi(ctx.Param("serverId"))
	if err != nil {
		ctx.JSON(400, gin.H{"error": "Must include valid ID of server"})
		return
	}

	var server data.Server
	if err := ctx.ShouldBindJSON(&server); err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid JSON", "trace": err.Error()})
		return
	}
	// Check if the ID in the URL matches the ID in the body (if provided)
	// If not, return a 400 Bad Request
	if server.Id != 0 && server.Id != uint(serverId) {
		ctx.JSON(400, gin.H{"error": "ID in URL does not match ID in body"})
		return
	}

	err = server.Update(sc.Database.Pool)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Unable to update Server", "trace": err.Error()})
		return
	}
	ctx.Header("HX-Redirect", "/servers/"+strconv.Itoa(int(server.Id))+"/view")
	ctx.JSON(200, server)
}
