package htmx

import (
	"kukus/nam/v2/layers/data"
	services "kukus/nam/v2/layers/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type TopologyNodeController struct {
	Database        *data.Database
	TopologyService services.TopologyNodeService
}

func NewTopologyNodeController(database *data.Database) TopologyNodeController {
	return TopologyNodeController{
		Database: database,
		TopologyService: services.TopologyNodeService{
			Database: database,
		},
	}
}

func (tnc TopologyNodeController) Init(routeGroup *gin.RouterGroup) {
	routeGroup.GET("/:id", tnc.RenderTopologyNode)
}

// Fancy endpoint, renders any topology node... almost... hopefully
// Never casted a struct to another struct this transitive way, crazy it works
func (tnc TopologyNodeController) RenderTopologyNode(ctx *gin.Context) {
	id := ctx.Param("id")
	tnId, err := strconv.Atoi(id)
	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{"error": "must be a valid topology node id!", "trace": err})
		return
	}
	tn, err := tnc.TopologyService.GetTopologyNodeById(uint(tnId))
	if err != nil {
		ctx.JSON(500, gin.H{"error": err})
		return
	}
	redirect := ""
	switch tn.Type {
	case "application_instance":
		redirect = "/applications"
	case "server":
		redirect = "/servers"
	case "healthcheck":
		redirect = "/healthchecks"
	}
	ctx.Redirect(301, "/htmx/"+redirect+"/"+id)
	return
}
