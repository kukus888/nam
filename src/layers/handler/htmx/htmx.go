package htmx

import (
	"kukus/nam/v2/layers/data"

	"github.com/gin-gonic/gin"
)

type HtmxController struct {
	Database *data.Database
}

func NewHtmxController(database *data.Database) HtmxController {
	return HtmxController{
		Database: database,
	}
}

func (hc HtmxController) Init(routeGroup *gin.RouterGroup) {
	NewApplicationView(hc.Database).Init(routeGroup.Group("/applications"))
	NewHtmxHealthHandler(hc.Database.Pool).Init(routeGroup.Group("/health"))
	NewHtmxActionHandler(hc.Database).Init(routeGroup.Group("/actions"))
	routeGroup.GET("/navbar_user", func(ctx *gin.Context) {
		user_id_uint64 := ctx.GetUint64("user_id")
		user, err := data.GetUserById(hc.Database.Pool, user_id_uint64)
		if err != nil {
			ctx.JSON(500, gin.H{"error": "unable to get user from database"})
			return
		}
		ctx.HTML(200, "component/navbar_user", gin.H{
			"Username": user.Username,
			"Color":    user.Color,
		})
	})
}
