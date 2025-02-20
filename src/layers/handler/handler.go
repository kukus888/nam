package handlers

import (
	"github.com/gin-gonic/gin"
)

// Generic HTTP 409 Method not allowed response
func MethodNotAllowed(ctx *gin.Context) {
	ctx.AbortWithStatus(409)
}

// Generic HTTP 501 Not Implemented response
func MethodNotImplemented(ctx *gin.Context) {
	ctx.AbortWithStatus(501)
}
