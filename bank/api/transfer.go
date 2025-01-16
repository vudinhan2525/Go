package api

import (
	"database/sql"
	"errors"
	"fmt"
	db "main/db/sqlc"
	"main/pkg/middlewares"
	"main/token"
	"net/http"

	"github.com/gin-gonic/gin"
)

type TransferReqBody struct {
	FromAccountId int64  `json:"from_account_id" binding:"required"`
	ToAccountId   int64  `json:"to_account_id" binding:"required"`
	Amount        int64  `json:"amount" binding:"required,gte=1"`
	Currency      string `json:"currency" binding:"required,currency"`
}

func (server *Server) transferMoney(ctx *gin.Context) {
	var req TransferReqBody
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	authPayload := ctx.MustGet(middlewares.AuthorizationPayloadKey).(*token.Payload)

	fromAcc, ok := server.checkValidAccount(ctx, req.FromAccountId, req.Currency)
	if fromAcc.Owner != int64(authPayload.UserID) {
		err := errors.New("from account doesn't belong to that user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return

	}
	if !ok {
		return
	}
	_, ok = server.checkValidAccount(ctx, req.FromAccountId, req.Currency)
	if !ok {
		return
	}
	result, err := server.Store.TransferTx(ctx, db.TransferTxParams{
		FromAccountId: req.FromAccountId,
		ToAccountId:   req.ToAccountId,
		Amount:        req.Amount,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "Create transfer successfully", "data": result})
}

func (server *Server) checkValidAccount(ctx *gin.Context, accountID int64, currency string) (db.Account, bool) {
	acc, err := server.Store.GetAccount(ctx, accountID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return acc, false
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return acc, false
	}

	if acc.Currency != currency {
		err := fmt.Errorf("account [%v] currency mismatch : %s vs %s", accountID, acc.Currency, currency)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return acc, false
	}
	return acc, true
}
