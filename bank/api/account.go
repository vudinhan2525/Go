package api

import (
	"database/sql"
	db "main/db/sqlc"
	"main/util"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CreateAccountParams struct {
	Owner    int64  `json:"owner" binding:"required"`
	Currency string `json:"currency" binding:"required,currency"`
}

func (server *Server) createAccount(ctx *gin.Context) {
	var req CreateAccountParams
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	acc, err := server.store.CreateAccount(ctx, db.CreateAccountParams{
		Owner:    req.Owner,
		Currency: req.Currency,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "Create account successfully", "data": acc})
}

type GetAccountParams struct {
	ID int64 `uri:"id" binding:"required,gte=1"`
}

func (server *Server) getAccountById(ctx *gin.Context) {
	var req GetAccountParams
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	acc, err := server.store.GetAccount(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "Get account successfully", "data": acc})
}

func (server *Server) getAccounts(ctx *gin.Context) {
	page, limit := util.GetPaginateFromRequest(ctx, "1", "5")

	accounts, err := server.store.ListAccounts(ctx, db.ListAccountsParams{
		Limit:  int32(limit),
		Offset: (page - 1) * limit,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "Get accounts list successfully", "data": accounts})
}
