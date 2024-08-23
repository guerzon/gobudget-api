package db

import (
	"context"

	"github.com/google/uuid"
)

// Database transaction for creating a user.
func (s *SQLStore) CreateUserTx(ctx context.Context, arg CreateUserParams, fn func(createdUser UserParams) error) (User, error) {

	var txUser UserTxResult

	txErr := s.execTransaction(ctx, func(q *Queries) error {
		var err error
		// Create the user in the database
		txUser.User, err = q.CreateUser(ctx, arg)
		if err != nil {
			return err
		}
		// Run the function that creates a Redis task that
		// sends the verification email
		if err := fn(UserParams{
			Username: txUser.User.Username,
		}); err != nil {
			return err
		}
		return nil
	})

	return txUser.User, txErr
}

// Database transaction for updating a user
func (s *SQLStore) UpdateUserTx(ctx context.Context, arg UpdateUserParams, fn func(createdUser UserParams) error) (User, error) {

	var txUser UserTxResult

	txErr := s.execTransaction(ctx, func(q *Queries) error {
		var err error
		txUser.User, err = q.UpdateUser(ctx, arg)
		if err != nil {
			return err
		}
		// only call the task if we are changing the email
		if arg.Email.Valid && arg.EmailVerified.Valid {
			if err := fn(UserParams{
				Username: txUser.User.Username,
			}); err != nil {
				return err
			}
		}
		return nil
	})

	return txUser.User, txErr
}

// Database transaction for deleting a user
func (s *SQLStore) DeleteUserTx(ctx context.Context, userArg UserParams, budgetIds []uuid.UUID, fn func(deleteUser UserParams) error) error {

	txError := s.execTransaction(ctx, func(q *Queries) error {
		// Delete email verifications
		if err := q.DeleteVerifyEmails(ctx, userArg.Username); err != nil {
			return err
		}
		// Delete sessions
		if err := q.DeleteUserSessions(ctx, userArg.Username); err != nil {
			return err
		}

		for i := range budgetIds {
			// Delete payees
			// TODO
			// Delete the transactions
			// TODO
			// Delete category groups
			cg, err := q.GetCategoryGroupsByBudgetId(ctx, budgetIds[i])
			if err != nil {
				return err
			}
			cgIds := make([]uuid.UUID, len(cg))
			for c := range cg {
				cgIds[c] = cg[c].ID
			}
			if err := q.DeleteCategoryGroups(ctx, budgetIds[i]); err != nil {
				return err
			}
			// Delete categories
			for c := range cgIds {
				if err := q.DeleteCategories(ctx, cgIds[c]); err != nil {
					return err
				}
			}
			// Delete the accounts
			if err := q.DeleteAccounts(ctx, budgetIds[i]); err != nil {
				return err
			}
		}
		if err := q.DeleteBudgets(ctx, userArg.Username); err != nil {
			return err
		}
		if err := fn(userArg); err != nil {
			return err
		}
		return nil
	})
	return txError
}
