package db

import (
	"context"
	"database/sql"
	"fmt"
)

type Store struct {
	*Queries
	db *sql.DB
}

type txKeyType struct{}

var txKey = txKeyType{}

func NewStore(db *sql.DB) *Store {
	return &Store{
		db:      db,
		Queries: New(db),
	}
}

func (store *Store) execTx(ctx context.Context, fn func(*Queries) error) error {
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

func (store *Store) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	var result TransferTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		txValue := ctx.Value(txKey)

		fmt.Println(txValue, "create transfer")
		result.transfer, err = q.CreateTransfer(ctx, CreateTransferParams{
			FromAccountID: arg.fromAccountId,
			ToAccountID:   arg.toAccountId,
			Amount:        arg.amount,
		})

		if err != nil {
			return err
		}

		fmt.Println(txValue, "create entry1")
		result.fromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			Amount:    -arg.amount,
			AccountID: arg.fromAccountId,
		})
		if err != nil {
			return err
		}

		fmt.Println(txValue, "create entry2")
		result.toEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			Amount:    arg.amount,
			AccountID: arg.toAccountId,
		})
		if err != nil {
			return err
		}

		fmt.Println(txValue, "get acc1")
		fromAccount, err := q.GetAccountForUpdate(ctx, arg.fromAccountId)
		if err != nil {
			return err
		}

		fmt.Println(txValue, "update acc1")
		newFromBalance := fromAccount.Balance - arg.amount
		result.fromAccount, err = q.UpdateAccount(ctx, UpdateAccountParams{
			ID:      arg.fromAccountId,
			Balance: newFromBalance,
		})
		if err != nil {
			return err
		}

		fmt.Println(txValue, "get acc2")
		toAccount, err := q.GetAccountForUpdate(ctx, arg.toAccountId)
		if err != nil {
			return err
		}

		fmt.Println(txValue, "update acc2")
		newToBalance := toAccount.Balance + arg.amount
		result.toAccount, err = q.UpdateAccount(ctx, UpdateAccountParams{
			ID:      arg.toAccountId,
			Balance: newToBalance,
		})
		if err != nil {
			return err
		}
		return nil
	})

	return result, err
}
