package api

import (
	"database/sql"
	"fmt"
	db "main/db/sqlc"
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
	if !server.checkValidCurrency(ctx, req.FromAccountId, req.Currency) {
		return
	}
	if !server.checkValidCurrency(ctx, req.ToAccountId, req.Currency) {
		return
	}
	result, err := server.store.TransferTx(ctx, db.TransferTxParams{
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

func (server *Server) checkValidCurrency(ctx *gin.Context, accountID int64, currency string) bool {
	acc, err := server.store.GetAccount(ctx, accountID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return false
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return false
	}

	if acc.Currency != currency {
		err := fmt.Errorf("account [%v] currency mismatch : %s vs %s", accountID, acc.Currency, currency)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return false
	}
	return true
}
