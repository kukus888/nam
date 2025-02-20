package htmx

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

func (hc HtmxController) Server(ctx *gin.Context) {
	serverId, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.AbortWithStatus(400)
		return
	}
	server, err := hc.ServerService.GetServerById(uint(serverId))
	if err != nil {
		ctx.AbortWithStatus(500)
		return
	}
	ctx.HTML(200, "template/server.small", gin.H{
		"Server": server,
	})
}
