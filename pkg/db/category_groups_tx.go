package db

import (
	"context"

	"github.com/google/uuid"
)

// Database transaction for deleting category group
func (s *SQLStore) DeleteCategoryGroupTx(ctx context.Context, categoryGroupId uuid.UUID) error {

	txErr := s.execTransaction(ctx, func(q *Queries) error {

		// Delete the categories
		if err := q.DeleteCategories(ctx, categoryGroupId); err != nil {
			return err
		}
		// Delete the category group
		if err := q.DeleteCategoryGroup(ctx, categoryGroupId); err != nil {
			return err
		}
		return nil
	})
	return txErr
}
