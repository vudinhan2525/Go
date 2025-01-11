package db

import (
	"context"
	"database/sql"
	"fmt"
)

type Store interface {
	Querier
	TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error)
}
type StoreSQL struct {
	*Queries
	db *sql.DB
}

type txKeyType struct{}

var txKey = txKeyType{}

func NewStore(db *sql.DB) Store {
	return &StoreSQL{
		db:      db,
		Queries: New(db),
	}
}

func (store *StoreSQL) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	q := New(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("rx error %v, rb error %v", err, rbErr)
		}
		return err
	}
	return tx.Commit()
}

type TransferTxParams struct {
	FromAccountId int64 `json:"from_account_id"`
	ToAccountId   int64 `json:"to_account_id"`
	Amount        int64 `json:"amount"`
}
type TransferTxResult struct {
	Transfer    Transfer `json:"transfer"`
	FromAccount Account  `json:"from_account"`
	ToAccount   Account  `json:"to_account"`
	FromEntry   Entry    `json:"from_entry"`
	ToEntry     Entry    `json:"to_entry"`
}

func (store *StoreSQL) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	var result TransferTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams{
			FromAccountID: arg.FromAccountId,
			ToAccountID:   arg.ToAccountId,
			Amount:        arg.Amount,
		})

		if err != nil {
			return err
		}

		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			Amount:    -arg.Amount,
			AccountID: arg.FromAccountId,
		})
		if err != nil {
			return err
		}

		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			Amount:    arg.Amount,
			AccountID: arg.ToAccountId,
		})
		if err != nil {
			return err
		}
		if arg.FromAccountId < arg.ToAccountId {
			result.FromAccount, err = q.UpdateAccountBalance(ctx, UpdateAccountBalanceParams{
				Amount: -arg.Amount,
				ID:     arg.FromAccountId,
			})
			if err != nil {
				return err
			}

			result.ToAccount, err = q.UpdateAccountBalance(ctx, UpdateAccountBalanceParams{
				Amount: arg.Amount,
				ID:     arg.ToAccountId,
			})
			if err != nil {
				return err
			}
		} else {
			result.ToAccount, err = q.UpdateAccountBalance(ctx, UpdateAccountBalanceParams{
				Amount: arg.Amount,
				ID:     arg.ToAccountId,
			})
			if err != nil {
				return err
			}
			result.FromAccount, err = q.UpdateAccountBalance(ctx, UpdateAccountBalanceParams{
				Amount: -arg.Amount,
				ID:     arg.FromAccountId,
			})
			if err != nil {
				return err
			}
		}

		return nil
	})

	return result, err
}
