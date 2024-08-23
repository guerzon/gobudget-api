package db

import (
	"context"

	"github.com/google/uuid"
)

// Database transaction for deleting a budget.
func (s *SQLStore) DeleteBudgetTx(ctx context.Context, budgetId uuid.UUID) error {

	txErr := s.execTransaction(ctx, func(q *Queries) error {

		// Delete payees
		// Delete the transactions
		// TODO
		// Delete category groups
		cg, err := q.GetCategoryGroupsByBudgetId(ctx, budgetId)
		if err != nil {
			return err
		}
		cgIds := make([]uuid.UUID, len(cg))
		for c := range cg {
			cgIds[c] = cg[c].ID
		}
		if err := q.DeleteCategoryGroups(ctx, budgetId); err != nil {
			return err
		}
		// Delete categories
		for c := range cgIds {
			if err := q.DeleteCategories(ctx, cgIds[c]); err != nil {
				return err
			}
		}
		// Delete the accounts
		if err := q.DeleteAccounts(ctx, budgetId); err != nil {
			return err
		}
		// Delete the budget
		if err := q.DeleteBudget(ctx, budgetId); err != nil {
			return err
		}

		return nil
	})
	return txErr
}
