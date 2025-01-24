package api

import (
	"database/sql"
	"errors"
	db "main/db/sqlc"
	"main/pkg/middlewares"
	"main/token"
	"main/util"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CreateAccountParams struct {
	Currency string `json:"currency" binding:"required,currency"`
}

func (server *Server) createAccount(ctx *gin.Context) {
	var req CreateAccountParams
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	authPayload := ctx.MustGet(middlewares.AuthorizationPayloadKey).(*token.Payload)
	acc, err := server.Store.CreateAccount(ctx, db.CreateAccountParams{
		Owner:    int64(authPayload.UserID),
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
	authPayload := ctx.MustGet(middlewares.AuthorizationPayloadKey).(*token.Payload)

	acc, err := server.Store.GetAccount(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	if acc.Owner != int64(authPayload.UserID) {
		err := errors.New("account doesn't belong to that user ")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "Get account successfully", "data": acc})
}

func (server *Server) getAccounts(ctx *gin.Context) {
	page, limit := util.GetPaginateFromRequest(ctx, "1", "5")
	authPayload := ctx.MustGet(middlewares.AuthorizationPayloadKey).(*token.Payload)

	accounts, err := server.Store.ListAccounts(ctx, db.ListAccountsParams{
		Owner:  int64(authPayload.UserID),
		Limit:  int32(limit),
		Offset: (page - 1) * limit,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "Get accounts list successfully", "data": accounts})
}
