package db

import "context"

type CreateUserTxParams struct {
	CreateUserParams CreateUserParams
	AfterCreate      func(user User) error
}
type CreateUserTxResult struct {
	User User
}

func (store *StoreSQL) CreateUserTx(ctx context.Context, arg CreateUserTxParams) (CreateUserTxResult, error) {
	var result CreateUserTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		user, err := q.CreateUser(ctx, arg.CreateUserParams)

		if err != nil {
			return err
		}

		err = arg.AfterCreate(user)

		if err != nil {
			return err
		}
		result = CreateUserTxResult{
			User: user,
		}
		return nil
	})

	return result, err
}
