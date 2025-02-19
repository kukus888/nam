/*
	This is only a template
*/

package handlers

import (
	data "kukus/nam/v2/layers/data"

	"github.com/gin-gonic/gin"
)

// Delete this, for imports only
var Db = data.ApplicationInstance{}

func Get(ctx *gin.Context) {
	ctx.AbortWithStatus(500)
}

func Post(ctx *gin.Context) {
	ctx.AbortWithStatus(500)
}

func Patch(ctx *gin.Context) {
	ctx.AbortWithStatus(500)
}

func Put(ctx *gin.Context) {
	ctx.AbortWithStatus(500)
}

func Delete(ctx *gin.Context) {
	ctx.AbortWithStatus(500)
}
