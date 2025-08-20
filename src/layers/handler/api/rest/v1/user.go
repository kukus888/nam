package v1

import (
	data "kukus/nam/v2/layers/data"

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
	// Create user in the database
}

func (h *UserHandler) UpdateUser(ctx *gin.Context) {
	// Handler logic for updating a user
}

func (h *UserHandler) DeleteUser(ctx *gin.Context) {
	// Handler logic for deleting a user
}
