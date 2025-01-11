package api

import (
	"database/sql"
	db "main/db/sqlc"
	"main/util"
	"net/http"

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
	acc, err := server.store.CreateUser(ctx, db.CreateUserParams{
		Email:          req.Email,
		FullName:       req.FullName,
		HashedPassword: password,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "Create users successfully", "data": acc})
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
	acc, err := server.store.GetUser(ctx, req.ID)
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
