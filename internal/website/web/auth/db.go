package auth

import (
	"context"
	"database/sql"
)

var DB *sql.DB

func Transact(ctx context.Context, fs ...func(*sql.Tx) error) (err error) {

	transaction, err := DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != context.Canceled {
			transaction.Rollback()
		}
	}()

	for _, f := range fs {
		if err := f(transaction); err != nil {
			return err
		}
	}

	return transaction.Commit()
}
