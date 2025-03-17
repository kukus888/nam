package handlers

import (
	"github.com/gin-gonic/gin"
)

// Generic HTTP 409 Method not allowed response
func MethodNotAllowed(ctx *gin.Context) {
	ctx.String(409, "Method not allowed")
	ctx.AbortWithStatus(409)
}

// Generic HTTP 501 Not Implemented response
func MethodNotImplemented(ctx *gin.Context) {
	ctx.String(501, "Not implemented")
	ctx.AbortWithStatus(501)
}
