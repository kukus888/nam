package v1

import (
	data "kukus/nam/v2/layers/data"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProfileHandler struct {
	Database *pgxpool.Pool
}

func NewProfileHandler(database *pgxpool.Pool) *ProfileHandler {
	return &ProfileHandler{
		Database: database,
	}
}

// Handler for changing the profile of the currently logged-in user
func (h *ProfileHandler) UpdateUser(ctx *gin.Context) {
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
	// Check, if the user is trying to assign himself something he isnt supposed to
	userFromContext, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(500, gin.H{"error": "Unable to get user from context"})
		return
	}
	if userFromContext != user.Id && role.Name != "admin" {
		ctx.JSON(403, gin.H{"error": "You are not allowed to change other users' profiles"})
		return
	}
	// Check, if he isnt trying to change other things
	existingUser, err := data.GetUserById(h.Database, user.Id)
	if err != nil {
		ctx.JSON(404, gin.H{"error": "User not found", "trace": err.Error()})
		return
	}
	if existingUser.RoleId != user.RoleId {
		ctx.JSON(403, gin.H{"error": "You are not allowed to change your role"})
		return
	}
	if existingUser.Email != user.Email {
		ctx.JSON(403, gin.H{"error": "You are not allowed to change your email"})
		return
	}
	if existingUser.Username != user.Username {
		ctx.JSON(403, gin.H{"error": "You are not allowed to change your username"})
		return
	}
	// Update user in the database
	if err := user.UpdateWithoutPassword(h.Database); err != nil {
		ctx.JSON(500, gin.H{"error": "Failed to update user", "trace": err.Error()})
		return
	}
	ctx.Header("HX-Redirect", "/profile")
	ctx.JSON(204, nil)
}

// Handler logic for changing the password of the currently logged-in user
func (h *ProfileHandler) UpdatePassword(ctx *gin.Context) {
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
	user, err := data.GetUserById(h.Database, userDelta.Id)
	if err != nil {
		ctx.JSON(404, gin.H{"error": "User not found", "trace": err.Error()})
		return
	}
	// Check, if the user is trying to assign himself something he isnt supposed to
	userFromContext, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(500, gin.H{"error": "Unable to get user from context"})
		return
	}
	if userFromContext != user.Id {
		ctx.JSON(403, gin.H{"error": "You are not allowed to change other users' passwords"})
		return
	}
	// Update user password in the database
	if err := user.UpdatePassword(h.Database, userDelta.Password); err != nil {
		ctx.JSON(500, gin.H{"error": "Failed to update user password", "trace": err.Error()})
		return
	}
	ctx.Header("HX-Redirect", "/profile")
	ctx.JSON(204, nil)
}
