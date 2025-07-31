package handlers

import (
	"github.com/gin-gonic/gin"
)

// Generic HTTP 401 Unauthorized response
func Unauthorized(ctx *gin.Context) {
	ctx.HTML(401, "pages/401", gin.H{})
	ctx.Abort()
}

// Generic HTTP 403 Forbidden response
func Forbidden(ctx *gin.Context) {
	ctx.HTML(403, "pages/403", gin.H{})
	ctx.Abort()
}

// Generic HTTP 404 Not found response
func NotFound(ctx *gin.Context) {
	ctx.HTML(404, "pages/404", gin.H{})
	ctx.Abort()
}

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
