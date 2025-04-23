package user

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	service *UserService
	logger  *log.Logger
}

func NewUserController(service *UserService, logger *log.Logger) *UserController {
	return &UserController{
		service: service,
		logger:  logger,
	}
}

func (u *UserController) CreateUser(ctx *gin.Context) {
	type CreateUserRequest struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	var user CreateUserRequest
	if err := ctx.ShouldBindJSON(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	reqCtx := ctx.Request.Context()
	reqCtx, cancel := context.WithTimeout(reqCtx, 5*time.Second)
	defer cancel()

	createdUser, err := u.service.CreateUser(reqCtx, user.Username, user.Email, user.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, createdUser)
}

func (u *UserController) GetAllUsers(ctx *gin.Context) {
	reqCtx := ctx.Request.Context()
	reqCtx, cancel := context.WithTimeout(reqCtx, 5*time.Second)
	defer cancel()

	users, err := u.service.GetAllUsers(reqCtx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, users)
}

func (u *UserController) GetUserByID(ctx *gin.Context) {
	id := ctx.Param("id")
	userID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user id"})
		return
	}

	reqCtx := ctx.Request.Context()
	reqCtx, cancel := context.WithTimeout(reqCtx, 5*time.Second)
	defer cancel()

	user, err := u.service.GetUserByID(reqCtx, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, user)
}

func (u *UserController) GetUserAccounts(ctx *gin.Context) {
	id := ctx.Param("id")
	userID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user id"})
		return
	}

	reqCtx := ctx.Request.Context()
	reqCtx, cancel := context.WithTimeout(reqCtx, 5*time.Second)
	defer cancel()

	accounts, err := u.service.GetUserAccounts(reqCtx, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, accounts)
}

func (u *UserController) UpdateUser(ctx *gin.Context) {
	id := ctx.Param("id")
	userID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user id"})
		return
	}

	type UpdateUserRequest struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	var user UpdateUserRequest
	if err := ctx.ShouldBindJSON(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	reqCtx := ctx.Request.Context()
	reqCtx, cancel := context.WithTimeout(reqCtx, 5*time.Second)
	defer cancel()

	err = u.service.UpdateUser(reqCtx, userID, user.Username, user.Email, user.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}

func (u *UserController) RegisterRoutes(router *gin.Engine) {
	group := router.Group("/users")
	{
		group.POST("", u.CreateUser)
		group.GET("/:id", u.GetUserByID)
		group.GET("/:id/accounts", u.GetUserAccounts)
		group.GET("", u.GetAllUsers)
		group.PUT("/:id", u.UpdateUser)
	}
}
