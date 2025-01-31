package api

import (
	"database/sql"
	db "main/db/sqlc"
	"main/pkg/middlewares"
	"main/token"
	"main/util"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type CreateUserRequest struct {
	FullName string `json:"fullname" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
}

func (server *Server) createUser(ctx *gin.Context) {
	var req CreateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	password, err := util.HashPassword(req.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	user, err := server.Store.CreateUser(ctx, db.CreateUserParams{
		Email:          req.Email,
		FullName:       req.FullName,
		HashedPassword: password,
		Role:           "user",
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "Create users successfully", "data": user})
}

type GetUserParams struct {
	ID int64 `uri:"id" binding:"required,gte=1"`
}

func (server *Server) getUsertById(ctx *gin.Context) {
	var req GetUserParams
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	acc, err := server.Store.GetUser(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "Get user successfully", "data": acc})
}

type LoginUserRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
}

func (server *Server) loginUser(ctx *gin.Context) {

	var req LoginUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(util.NewValidationError(err, "invalid request body"))
		return
	}
	user, err := server.Store.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.Error(util.NewNotFoundError(err, "doesn't have user with this email"))
			return
		}
		ctx.Error(util.NewInternalServerError(err, err.Error()))
		return
	}

	err = util.VerifyPassword(req.Password, user.HashedPassword)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	accessToken, _, err := server.TokenMaker.CreateToken(strconv.FormatInt(user.UserID, 10), user.Email, user.Role, server.Config.TokenDuration)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	refreshToken, refreshPayload, err := server.TokenMaker.CreateToken(strconv.FormatInt(user.UserID, 10), user.Email, user.Role, server.Config.RefreshTokenDuration)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	_, err = server.Store.CreateSession(ctx, db.CreateSessionParams{
		ID:           refreshPayload.ID,
		UserID:       int64(refreshPayload.UserID),
		RefreshToken: refreshToken,
		UserAgent:    "",
		ClientIp:     "",
		IsBlocked:    false,
		ExpiredAt:    refreshPayload.ExpiredAt,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"status": "Login successfully", "data": user, "access_token": accessToken, "refresh_token": refreshToken,
	})
}

type UpdateUserRequest struct {
	Email    string `json:"email"`
	FullName string `json:"fullname"`
	Password string `json:"password"`
}

func (server *Server) updateUser(ctx *gin.Context) {
	var req UpdateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(util.NewValidationError(err, "invalid request body"))
		return
	}
	authPayload := ctx.MustGet(middlewares.AuthorizationPayloadKey).(*token.Payload)
	params := db.UpdateUserParams{
		Email:    sql.NullString{String: req.Email, Valid: req.Email != ""},
		FullName: sql.NullString{String: req.FullName, Valid: req.FullName != ""},
		UserID:   int64(authPayload.UserID),
	}
	if req.Password != "" {
		hashedPassword, err := util.HashPassword(req.Password)
		if err != nil {
			ctx.Error(util.NewInternalServerError(err, "error when hash password"))
			return
		}
		params.HashedPassword = sql.NullString{String: hashedPassword, Valid: hashedPassword != ""}
		params.PasswordChangedAt = sql.NullTime{Time: time.Now(), Valid: hashedPassword != ""}
	}
	user, err := server.Store.UpdateUser(ctx, params)
	if err != nil {
		ctx.Error(util.NewInternalServerError(err, "error when update user"))
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "Login successfully", "data": user,
	})
}
