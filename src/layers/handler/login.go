package handlers

import (
	"fmt"
	"hash/fnv"
	"kukus/nam/v2/layers/data"
	services "kukus/nam/v2/layers/service"
	"log/slog"
	"slices"

	"github.com/gin-gonic/gin"
)

type LoginPageHandler struct {
	Database *data.Database
	Logger   *slog.Logger
}

func NewLoginPageHandler(database *data.Database, logger *slog.Logger) LoginPageHandler {
	return LoginPageHandler{
		Database: database,
		Logger:   logger,
	}
}

func (pc LoginPageHandler) Init(routeGroup *gin.RouterGroup) {
	routeGroup.GET("/", func(ctx *gin.Context) {
		ctx.HTML(200, "pages/login", gin.H{})
	})
	routeGroup.POST("/", func(ctx *gin.Context) {
		var User data.UserLoginDTO
		if err := ctx.ShouldBindJSON(&User); err != nil {
			pc.Logger.Warn("Invalid login attempt", "error", err, "trace", err.Error())
			ctx.JSON(400, gin.H{"error": "Invalid input"})
			return
		}
		if *User.Username == "" || *User.Password == "" {
			pc.Logger.Warn("Login attempt with missing username or password")
			ctx.JSON(400, gin.H{"error": "Username and password are required"})
			return
		}
		userDao, err := data.GetUserByUsername(pc.Database.Pool, *User.Username)
		if err != nil {
			pc.Logger.Error("Database error during login", "error", err, "trace", err.Error(), "username", *User.Username)
			ctx.JSON(500, gin.H{"error": "Database error", "trace": err.Error()})
			return
		}
		if userDao == nil {
			pc.Logger.Warn("User not found", "username", *User.Username)
			ctx.JSON(401, gin.H{"error": "Invalid credentials"})
			return
		}
		if err := data.VerifyPassword(userDao.PasswordHash, *User.Password); err != nil {
			pc.Logger.Warn("Invalid login attempt", "username", *User.Username)
			ctx.JSON(401, gin.H{"error": "Invalid credentials"})
			return
		}
		// Set JWT token
		token, err := services.GenerateToken(*userDao)
		if err != nil {
			pc.Logger.Error("Error generating token", "error", err, "trace", err.Error(), "username", *User.Username)
			ctx.JSON(500, gin.H{"error": "Error generating token", "trace": err.Error()})
			return
		}
		h := fnv.New64()
		h.Write([]byte(token))
		pc.Logger.Info("User logged in successfully", "username", *User.Username, "session", fmt.Sprintf("%x", h.Sum64()))
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
			pc.Logger.Warn("Attempt to access setup route when users already exist", "error", err, "trace", err.Error())
			ctx.JSON(400, gin.H{"error": "Admin user already exists"})
			return
		}
		var User data.UserDTO
		if err := ctx.ShouldBindJSON(&User); err != nil {
			pc.Logger.Warn("Invalid input for admin user setup", "error", err, "trace", err.Error())
			ctx.JSON(400, gin.H{"error": "Invalid input"})
			return
		}
		roles, err := data.GetAllRoles(pc.Database.Pool)
		if err != nil {
			pc.Logger.Error("Error fetching roles for admin user setup", "error", err, "trace", err.Error())
			ctx.JSON(500, gin.H{"error": "Unable to get role list", "trace": err.Error()})
			return
		}
		adminRoleId := slices.IndexFunc(roles, func(r data.Role) bool {
			return r.Name == "Admin"
		})
		id, err := data.CreateUser(pc.Database.Pool, User, uint64(roles[adminRoleId].Id))
		if err != nil {
			pc.Logger.Error("Error creating admin user", "error", err, "trace", err.Error())
			ctx.JSON(500, gin.H{"error": "Unable to create user", "trace": err.Error()})
			return
		}
		ctx.JSON(200, gin.H{"message": "Admin user created successfully", "user_id": id})
	})
}
