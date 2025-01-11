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
	fromAccountId int64 "json:from_account_id"
	toAccountId   int64 "json:to_account_id"
	amount        int64 "json:amount"
}
type TransferTxResult struct {
	transfer    Transfer
	fromAccount Account
	toAccount   Account
	fromEntry   Entry
	toEntry     Entry
}

func (store *StoreSQL) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	var result TransferTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		result.transfer, err = q.CreateTransfer(ctx, CreateTransferParams{
			FromAccountID: arg.fromAccountId,
			ToAccountID:   arg.toAccountId,
			Amount:        arg.amount,
		})

		if err != nil {
			return err
		}

		result.fromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			Amount:    -arg.amount,
			AccountID: arg.fromAccountId,
		})
		if err != nil {
			return err
		}

		result.toEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			Amount:    arg.amount,
			AccountID: arg.toAccountId,
		})
		if err != nil {
			return err
		}
		if arg.fromAccountId < arg.toAccountId {
			result.fromAccount, err = q.UpdateAccountBalance(ctx, UpdateAccountBalanceParams{
				Amount: -arg.amount,
				ID:     arg.fromAccountId,
			})
			if err != nil {
				return err
			}

			result.toAccount, err = q.UpdateAccountBalance(ctx, UpdateAccountBalanceParams{
				Amount: arg.amount,
				ID:     arg.toAccountId,
			})
			if err != nil {
				return err
			}
		} else {
			result.toAccount, err = q.UpdateAccountBalance(ctx, UpdateAccountBalanceParams{
				Amount: arg.amount,
				ID:     arg.toAccountId,
			})
			if err != nil {
				return err
			}
			result.fromAccount, err = q.UpdateAccountBalance(ctx, UpdateAccountBalanceParams{
				Amount: -arg.amount,
				ID:     arg.fromAccountId,
			})
			if err != nil {
				return err
			}
		}

		return nil
	})

	return result, err
}
