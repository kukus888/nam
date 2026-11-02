package handlers

import (
	"context"
	"fmt"
	"hash/fnv"
	"kukus/nam/v2/layers/data"
	services "kukus/nam/v2/layers/service"
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
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
		errCode := ctx.Query("err")
		errorMessage := ""
		switch errCode {
		case "missing_credentials":
			errorMessage = "Please enter both username and password."
		case "invalid_credentials":
			errorMessage = "We couldn't sign you in. Check your username and password, then try again."
		case "temporary_issue":
			errorMessage = "We couldn't sign you in right now. Please try again in a moment."
		}

		ctx.HTML(200, "pages/login", gin.H{"Error": errorMessage})
	})
	routeGroup.POST("/", func(ctx *gin.Context) {
		username := ctx.PostForm("username")
		password := ctx.PostForm("password")

		// Keep JSON support for API-like callers while preferring standard form posts.
		if username == "" && password == "" {
			var user data.UserLoginDTO
			if err := ctx.ShouldBindJSON(&user); err == nil {
				if user.Username != nil {
					username = *user.Username
				}
				if user.Password != nil {
					password = *user.Password
				}
			}
		}

		if username == "" || password == "" {
			pc.Logger.Warn("Login attempt with missing username or password")
			ctx.Redirect(303, "/login?err=missing_credentials")
			return
		}

		userDao, err := data.GetUserByUsername(pc.Database.Pool, username)
		if err != nil {
			pc.Logger.Error("Database error during login", "error", err, "trace", err.Error(), "username", username)
			ctx.Redirect(303, "/login?err=temporary_issue")
			return
		}
		if userDao == nil {
			pc.Logger.Warn("User not found", "username", username)
			ctx.Redirect(303, "/login?err=invalid_credentials")
			return
		}
		if err := data.VerifyPassword(userDao.PasswordHash, password); err != nil {
			pc.Logger.Warn("Invalid login attempt", "username", username)
			ctx.Redirect(303, "/login?err=invalid_credentials")
			return
		}
		// Set JWT token
		token, err := services.GenerateToken(*userDao)
		if err != nil {
			pc.Logger.Error("Error generating token", "error", err, "trace", err.Error(), "username", username)
			ctx.Redirect(303, "/login?err=temporary_issue")
			return
		}
		h := fnv.New64()
		h.Write([]byte(token))
		pc.Logger.Info("User logged in successfully", "username", username, "session", fmt.Sprintf("%x", h.Sum64()))
		// Store the token in a cookie or return it in the response
		ctx.Header("Set-Cookie", "token="+token+"; Path=/; HttpOnly; Secure; SameSite=Strict")
		ctx.Redirect(303, "/dashboards")
	})
	// One time only setup route, to generate admin user
	routeGroup.POST("/setup", func(ctx *gin.Context) {
		var User data.UserDTO
		if err := ctx.ShouldBindJSON(&User); err != nil {
			pc.Logger.Warn("Invalid input for admin user setup", "error", err, "trace", err.Error())
			ctx.JSON(400, gin.H{"error": "Invalid input"})
			return
		}

		tx, err := pc.Database.Pool.BeginTx(context.Background(), pgx.TxOptions{})
		if err != nil {
			pc.Logger.Error("Error starting transaction for admin user setup", "error", err, "trace", err.Error())
			ctx.JSON(500, gin.H{"error": "Unable to initialize setup", "trace": err.Error()})
			return
		}
		defer tx.Rollback(context.Background())

		_, err = tx.Exec(context.Background(), `SELECT pg_advisory_xact_lock($1);`, int64(273451))
		if err != nil {
			pc.Logger.Error("Error locking setup route", "error", err, "trace", err.Error())
			ctx.JSON(500, gin.H{"error": "Unable to initialize setup", "trace": err.Error()})
			return
		}

		var userCount int
		err = tx.QueryRow(context.Background(), `SELECT COUNT(*) FROM "user";`).Scan(&userCount)
		if err != nil {
			pc.Logger.Error("Error checking existing users during setup", "error", err, "trace", err.Error())
			ctx.JSON(500, gin.H{"error": "Unable to initialize setup", "trace": err.Error()})
			return
		}
		if userCount > 0 {
			pc.Logger.Warn("Attempt to access setup route when users already exist")
			ctx.JSON(400, gin.H{"error": "Admin user already exists"})
			return
		}

		var adminRoleId uint64
		err = tx.QueryRow(context.Background(), `SELECT id FROM "role" WHERE name = 'Admin' LIMIT 1;`).Scan(&adminRoleId)
		if err != nil {
			pc.Logger.Error("Error fetching admin role for setup", "error", err, "trace", err.Error())
			ctx.JSON(500, gin.H{"error": "Unable to get admin role", "trace": err.Error()})
			return
		}

		hash, err := data.HashPassword(User.Password)
		if err != nil {
			pc.Logger.Error("Error hashing password for admin user setup", "error", err, "trace", err.Error())
			ctx.JSON(500, gin.H{"error": "Unable to create user", "trace": err.Error()})
			return
		}

		var id uint64
		err = tx.QueryRow(context.Background(), `
			INSERT INTO "user" (username, password_hash, email, role_id)
			VALUES ($1, $2, $3, $4)
			RETURNING id;
		`, User.Username, hash, User.Email, adminRoleId).Scan(&id)
		if err != nil {
			pc.Logger.Error("Error creating admin user", "error", err, "trace", err.Error())
			ctx.JSON(500, gin.H{"error": "Unable to create user", "trace": err.Error()})
			return
		}

		if err := tx.Commit(context.Background()); err != nil {
			pc.Logger.Error("Error committing admin user setup", "error", err, "trace", err.Error())
			ctx.JSON(500, gin.H{"error": "Unable to finalize setup", "trace": err.Error()})
			return
		}

		ctx.JSON(200, gin.H{"message": "Admin user created successfully", "user_id": id})
	})
}
