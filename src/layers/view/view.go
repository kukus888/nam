package view

import (
	"kukus/nam/v2/layers/data"

	"github.com/gin-gonic/gin"
)

type Node interface {
	RenderTopologyNode(*data.TopologyNode, *gin.Context)
}
