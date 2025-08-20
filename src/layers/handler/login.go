package handlers

import (
	"kukus/nam/v2/layers/data"
	services "kukus/nam/v2/layers/service"
	"slices"

	"github.com/gin-gonic/gin"
)

type LoginPageHandler struct {
	Database *data.Database
}

func NewLoginPageHandler(database *data.Database) LoginPageHandler {
	return LoginPageHandler{
		Database: database,
	}
}

func (pc LoginPageHandler) Init(routeGroup *gin.RouterGroup) {
	routeGroup.GET("/", func(ctx *gin.Context) {
		ctx.HTML(200, "pages/login", gin.H{})
	})
	routeGroup.POST("/", func(ctx *gin.Context) {
		var User data.UserLoginDTO
		if err := ctx.ShouldBindJSON(&User); err != nil {
			ctx.JSON(400, gin.H{"error": "Invalid input"})
			return
		}
		if *User.Username == "" || *User.Password == "" {
			ctx.JSON(400, gin.H{"error": "Username and password are required"})
			return
		}
		userDao, err := data.GetUserByUsername(pc.Database.Pool, *User.Username)
		if err != nil {
			ctx.JSON(500, gin.H{"error": "Database error", "trace": err.Error()})
			return
		}
		if userDao == nil {
			ctx.JSON(404, gin.H{"error": "User not found"})
			return
		}
		if err := data.VerifyPassword(userDao.PasswordHash, *User.Password); err != nil {
			ctx.JSON(401, gin.H{"error": "Invalid credentials"})
			return
		}
		// TODO: JWT token
		// TODO: Set Authorization header with token
		token, err := services.GenerateToken(*userDao)
		if err != nil {
			ctx.JSON(500, gin.H{"error": "Error generating token", "trace": err.Error()})
			return
		}
		// Store the token in a cookie or return it in the response
		ctx.Header("Set-Cookie", "token="+token+"; Path=/; HttpOnly; Secure; SameSite=Strict")
		// Redirect to dashboard
		ctx.Header("HX-Redirect", "/dashboard")
		ctx.String(200, "Login successful, redirecting to dashboard")
	})
	// One time only setup route, to generate admin user
	routeGroup.POST("/setup", func(ctx *gin.Context) {
		// This route should only be used once to create the initial admin user
		if count, err := data.GetUserCount(pc.Database.Pool); count > 0 || err != nil {
			ctx.JSON(400, gin.H{"error": "Admin user already exists"})
			return
		}
		var User data.UserDTO
		if err := ctx.ShouldBindJSON(&User); err != nil {
			ctx.JSON(400, gin.H{"error": "Invalid input"})
			return
		}
		roles, err := data.GetAllRoles(pc.Database.Pool)
		if err != nil {
			ctx.JSON(500, gin.H{"error": "Unable to get role list", "trace": err.Error()})
			return
		}
		adminRoleId := slices.IndexFunc(roles, func(r data.Role) bool {
			return r.Name == "Admin"
		})
		id, err := data.CreateUser(pc.Database.Pool, User, uint64(roles[adminRoleId].Id))
		if err != nil {
			ctx.JSON(500, gin.H{"error": "Unable to create user", "trace": err.Error()})
			return
		}
		ctx.JSON(200, gin.H{"message": "Admin user created successfully", "user_id": id})
	})
}
