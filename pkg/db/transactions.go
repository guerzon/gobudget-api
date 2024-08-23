package db

import (
	"context"
	"fmt"
)

// Reusable function for executing database transactions.
func (store *SQLStore) execTransaction(ctx context.Context, fn func(*Queries) error) error {

	// Start a transaction
	trans, err := store.connPool.Begin(ctx)
	if err != nil {
		return err
	}

	// Get a Queries object
	query := New(trans)

	// Call the input function with the query obj
	err = fn(query)
	if err != nil {
		// if there is an error, roll back
		if rollbackErr := trans.Rollback(ctx); rollbackErr != nil {
			return fmt.Errorf("transaction error: %v, rollback error: %v", err, rollbackErr)
		}
		return err
	}

	return trans.Commit(ctx)
}
