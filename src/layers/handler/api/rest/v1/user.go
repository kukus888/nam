package v1

import (
	data "kukus/nam/v2/layers/data"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserHandler struct {
	Database *pgxpool.Pool
}

func NewUserHandler(database *pgxpool.Pool) *UserHandler {
	return &UserHandler{
		Database: database,
	}
}

// Handler logic for creating a user
func (h *UserHandler) CreateUser(ctx *gin.Context) {
	var user data.UserDTO
	if err := ctx.ShouldBindBodyWithJSON(&user); err != nil {
		ctx.JSON(400, gin.H{"error": "Unable to bind JSON data", "trace": err.Error()})
		return
	}
	// Check data
	if user.Username == "" || user.Email == "" || user.Password == "" {
		ctx.JSON(400, gin.H{"error": "Invalid username, email or password"})
		return
	}
	// Find corresponding role
	role, err := data.GetRoleById(h.Database, int(user.RoleId))
	if role == nil || err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid role ID", "trace": err.Error()})
		return
	}
	// Create user in the database
	userId, err := data.CreateUser(h.Database, user, role.Id)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Failed to create user", "trace": err.Error()})
		return
	}
	ctx.Header("HX-Redirect", "/settings/users")
	ctx.JSON(201, gin.H{"user_id": userId})
}

func (h *UserHandler) UpdateUser(ctx *gin.Context) {
	var user data.User
	if err := ctx.ShouldBindBodyWithJSON(&user); err != nil {
		ctx.JSON(400, gin.H{"error": "Unable to bind JSON data", "trace": err.Error()})
		return
	}
	// Check data
	if user.Id == 0 || user.Username == "" || user.Email == "" || user.RoleId == 0 {
		ctx.JSON(400, gin.H{"error": "Invalid user ID, username, email or role ID"})
		return
	}
	// Find corresponding role
	role, err := data.GetRoleById(h.Database, int(user.RoleId))
	if role == nil || err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid role ID", "trace": err.Error()})
		return
	}
	// Update user in the database
	if err := user.UpdateWithoutPassword(h.Database); err != nil {
		ctx.JSON(500, gin.H{"error": "Failed to update user", "trace": err.Error()})
		return
	}
	ctx.Header("HX-Redirect", "/settings/users")
	ctx.JSON(204, nil)
}

func (h *UserHandler) UpdatePassword(ctx *gin.Context) {
	var userDelta data.UserChangePasswordDTO
	if err := ctx.ShouldBindBodyWithJSON(&userDelta); err != nil {
		ctx.JSON(400, gin.H{"error": "Unable to bind JSON data", "trace": err.Error()})
		return
	}
	// Check data
	if userDelta.Id == 0 || userDelta.Password == "" {
		ctx.JSON(400, gin.H{"error": "Invalid user ID or password"})
		return
	}
	user, err := data.GetUserById(h.Database, int(userDelta.Id))
	if err != nil {
		ctx.JSON(404, gin.H{"error": "User not found", "trace": err.Error()})
		return
	}
	// Update user password in the database
	if err := user.UpdatePassword(h.Database, userDelta.Password); err != nil {
		ctx.JSON(500, gin.H{"error": "Failed to update user password", "trace": err.Error()})
		return
	}
	ctx.Header("HX-Redirect", "/settings/users")
	ctx.JSON(204, nil)
}

func (h *UserHandler) DeleteUser(ctx *gin.Context) {
	idParam := ctx.Param("id")
	userId, err := strconv.Atoi(idParam)
	if err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid user ID", "trace": err.Error()})
		return
	}
	user, err := data.GetUserById(h.Database, userId)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Failed to get user", "trace": err.Error()})
		return
	}
	if err = user.Delete(h.Database); err != nil {
		ctx.JSON(500, gin.H{"error": "Failed to delete user", "trace": err.Error()})
		return
	}
	ctx.Header("HX-Redirect", "/settings/users")
	ctx.Status(204)
}
