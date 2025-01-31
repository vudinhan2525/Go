package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTxTransfer(t *testing.T) {
	store := NewStore(testDb)

	ac1 := createTestAccount(t)
	ac2 := createTestAccount(t)
	fmt.Println(">>>>before", ac1.Balance, ac2.Balance)
	amt := int64(30)

	errs := make(chan error)
	results := make(chan TransferTxResult)
	n := 20
	for i := 0; i < n; i++ {
		go func() {
			txValue := fmt.Sprint("transfer ", i+1)
			ctx := context.WithValue(context.Background(), txKey, txValue)
			result, err := store.TransferTx(ctx, TransferTxParams{
				FromAccountId: ac1.ID,
				ToAccountId:   ac2.ID,
				Amount:        amt,
			})
			errs <- err
			results <- result
		}()
	}
	used := make(map[int]bool)
	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)

		result := <-results
		require.NotEmpty(t, result)
		// check transfer
		transfer := result.Transfer
		require.NotEmpty(t, transfer)
		require.Equal(t, ac1.ID, transfer.FromAccountID)
		require.Equal(t, ac2.ID, transfer.ToAccountID)
		require.Equal(t, amt, transfer.Amount)
		require.NotZero(t, transfer.ID)
		require.NotZero(t, transfer.CreatedAt)

		_, err = store.GetTransfer(context.Background(), transfer.ID)
		require.NoError(t, err)

		// check entries
		fromEntry := result.FromEntry
		require.NotEmpty(t, fromEntry)
		require.Equal(t, ac1.ID, fromEntry.AccountID)
		require.Equal(t, -amt, fromEntry.Amount)
		require.NotZero(t, fromEntry.ID)
		require.NotZero(t, fromEntry.CreatedAt)

		_, err = store.GetEntry(context.Background(), fromEntry.ID)
		require.NoError(t, err)

		toEntry := result.ToEntry
		require.NotEmpty(t, toEntry)
		require.Equal(t, ac2.ID, toEntry.AccountID)
		require.Equal(t, amt, toEntry.Amount)
		require.NotZero(t, toEntry.ID)
		require.NotZero(t, toEntry.CreatedAt)

		_, err = store.GetEntry(context.Background(), toEntry.ID)
		require.NoError(t, err)

		// check accounts
		fromAccount := result.FromAccount
		require.NotEmpty(t, fromAccount)
		require.Equal(t, ac1.ID, fromAccount.ID)

		toAccount := result.ToAccount
		require.NotEmpty(t, toAccount)
		require.Equal(t, ac2.ID, toAccount.ID)

		fmt.Println(">>>>tx", fromAccount.Balance, toAccount.Balance)

		// check balance

		diff1 := ac1.Balance - fromAccount.Balance
		diff2 := toAccount.Balance - ac2.Balance
		require.Equal(t, diff1, diff2)
		require.True(t, diff1 > 0)
		require.True(t, diff1%amt >= 0)

		k := int(diff1 / amt)
		require.True(t, k >= 1 && k <= n)
		require.True(t, used[k] == false)
		used[k] = true
	}
	// check final balance
	updatedAccount1, err := testQueries.GetAccount(context.Background(), ac1.ID)
	require.NoError(t, err)
	updatedAccount2, err := testQueries.GetAccount(context.Background(), ac2.ID)
	require.NoError(t, err)

	fmt.Println(">>>>after", updatedAccount1.Balance, updatedAccount2.Balance)

	require.Equal(t, ac1.Balance-int64(n)*amt, updatedAccount1.Balance)
	require.Equal(t, ac2.Balance+int64(n)*amt, updatedAccount2.Balance)
}
func TestTxTransferDeadlock(t *testing.T) {
	store := NewStore(testDb)

	ac1 := createTestAccount(t)
	ac2 := createTestAccount(t)

	amt := int64(30)
	n := 20
	errs := make(chan error)
	for i := 0; i < n; i++ {
		fromAccount := ac1.ID
		toAccount := ac2.ID
		if i%2 == 0 {
			fromAccount = ac2.ID
			toAccount = ac1.ID
		}
		go func() {

			_, err := store.TransferTx(context.Background(), TransferTxParams{
				FromAccountId: fromAccount,
				ToAccountId:   toAccount,
				Amount:        amt,
			})
			errs <- err
		}()
	}
	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)
	}
	// check final balance
	updatedAccount1, err := testQueries.GetAccount(context.Background(), ac1.ID)
	require.NoError(t, err)
	updatedAccount2, err := testQueries.GetAccount(context.Background(), ac2.ID)
	require.NoError(t, err)

	require.Equal(t, ac1.Balance, updatedAccount1.Balance)
	require.Equal(t, ac2.Balance, updatedAccount2.Balance)
}
