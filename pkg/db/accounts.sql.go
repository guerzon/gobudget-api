// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: accounts.sql

package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

const createAccount = `-- name: CreateAccount :one
INSERT INTO accounts (
    budget_id,
    name,
    type,
    balance
) VALUES (
    $1, $2, $3, $4
) RETURNING id, budget_id, name, type, closed, note, balance, cleared_balance, uncleared_balance, last_reconciled_at
`

type CreateAccountParams struct {
	BudgetID uuid.UUID `json:"budget_id"`
	Name     string    `json:"name"`
	Type     string    `json:"type"`
	Balance  int32     `json:"balance"`
}

func (q *Queries) CreateAccount(ctx context.Context, arg CreateAccountParams) (Account, error) {
	row := q.db.QueryRow(ctx, createAccount,
		arg.BudgetID,
		arg.Name,
		arg.Type,
		arg.Balance,
	)
	var i Account
	err := row.Scan(
		&i.ID,
		&i.BudgetID,
		&i.Name,
		&i.Type,
		&i.Closed,
		&i.Note,
		&i.Balance,
		&i.ClearedBalance,
		&i.UnclearedBalance,
		&i.LastReconciledAt,
	)
	return i, err
}

const deleteAccount = `-- name: DeleteAccount :exec
DELETE FROM accounts WHERE id = $1
`

func (q *Queries) DeleteAccount(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.Exec(ctx, deleteAccount, id)
	return err
}

const deleteAccounts = `-- name: DeleteAccounts :exec
DELETE FROM accounts WHERE budget_id = $1
`

func (q *Queries) DeleteAccounts(ctx context.Context, budgetID uuid.UUID) error {
	_, err := q.db.Exec(ctx, deleteAccounts, budgetID)
	return err
}

const getAccount = `-- name: GetAccount :one
SELECT id, budget_id, name, type, closed, note, balance, cleared_balance, uncleared_balance, last_reconciled_at FROM accounts WHERE budget_id = $1 and id = $2
`

type GetAccountParams struct {
	BudgetID uuid.UUID `json:"budget_id"`
	ID       uuid.UUID `json:"id"`
}

func (q *Queries) GetAccount(ctx context.Context, arg GetAccountParams) (Account, error) {
	row := q.db.QueryRow(ctx, getAccount, arg.BudgetID, arg.ID)
	var i Account
	err := row.Scan(
		&i.ID,
		&i.BudgetID,
		&i.Name,
		&i.Type,
		&i.Closed,
		&i.Note,
		&i.Balance,
		&i.ClearedBalance,
		&i.UnclearedBalance,
		&i.LastReconciledAt,
	)
	return i, err
}

const getAccounts = `-- name: GetAccounts :many
SELECT id, budget_id, name, type, closed, note, balance, cleared_balance, uncleared_balance, last_reconciled_at FROM accounts WHERE budget_id = $1
`

func (q *Queries) GetAccounts(ctx context.Context, budgetID uuid.UUID) ([]Account, error) {
	rows, err := q.db.Query(ctx, getAccounts, budgetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Account{}
	for rows.Next() {
		var i Account
		if err := rows.Scan(
			&i.ID,
			&i.BudgetID,
			&i.Name,
			&i.Type,
			&i.Closed,
			&i.Note,
			&i.Balance,
			&i.ClearedBalance,
			&i.UnclearedBalance,
			&i.LastReconciledAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getBudgetAccount = `-- name: GetBudgetAccount :one
SELECT b.id, owner_username, b.name, currency_code, a.id, budget_id, a.name, type, closed, note, balance, cleared_balance, uncleared_balance, last_reconciled_at FROM budgets b, accounts a
WHERE b.id = a.budget_id and b.id = $1 and a.id = $2 and b.owner_username = $3
`

type GetBudgetAccountParams struct {
	ID            uuid.UUID `json:"id"`
	ID_2          uuid.UUID `json:"id_2"`
	OwnerUsername string    `json:"owner_username"`
}

type GetBudgetAccountRow struct {
	ID               uuid.UUID   `json:"id"`
	OwnerUsername    string      `json:"owner_username"`
	Name             string      `json:"name"`
	CurrencyCode     string      `json:"currency_code"`
	ID_2             uuid.UUID   `json:"id_2"`
	BudgetID         uuid.UUID   `json:"budget_id"`
	Name_2           string      `json:"name_2"`
	Type             string      `json:"type"`
	Closed           bool        `json:"closed"`
	Note             pgtype.Text `json:"note"`
	Balance          int32       `json:"balance"`
	ClearedBalance   int32       `json:"cleared_balance"`
	UnclearedBalance int32       `json:"uncleared_balance"`
	LastReconciledAt time.Time   `json:"last_reconciled_at"`
}

func (q *Queries) GetBudgetAccount(ctx context.Context, arg GetBudgetAccountParams) (GetBudgetAccountRow, error) {
	row := q.db.QueryRow(ctx, getBudgetAccount, arg.ID, arg.ID_2, arg.OwnerUsername)
	var i GetBudgetAccountRow
	err := row.Scan(
		&i.ID,
		&i.OwnerUsername,
		&i.Name,
		&i.CurrencyCode,
		&i.ID_2,
		&i.BudgetID,
		&i.Name_2,
		&i.Type,
		&i.Closed,
		&i.Note,
		&i.Balance,
		&i.ClearedBalance,
		&i.UnclearedBalance,
		&i.LastReconciledAt,
	)
	return i, err
}

const updateAccount = `-- name: UpdateAccount :one
UPDATE accounts
SET
    name = COALESCE($3, name),
    type = COALESCE($4, type),
    closed = COALESCE($5, closed),
    note = COALESCE($6, note),
    balance = COALESCE($7, balance),
    cleared_balance = COALESCE($8, cleared_balance),
    uncleared_balance = COALESCE($9, uncleared_balance),
    last_reconciled_at = COALESCE($10, last_reconciled_at)
WHERE id = $1 AND budget_id = $2
RETURNING id, budget_id, name, type, closed, note, balance, cleared_balance, uncleared_balance, last_reconciled_at
`

type UpdateAccountParams struct {
	ID               uuid.UUID          `json:"id"`
	BudgetID         uuid.UUID          `json:"budget_id"`
	Name             pgtype.Text        `json:"name"`
	Type             pgtype.Text        `json:"type"`
	Closed           pgtype.Bool        `json:"closed"`
	Note             pgtype.Text        `json:"note"`
	Balance          pgtype.Int4        `json:"balance"`
	ClearedBalance   pgtype.Int4        `json:"cleared_balance"`
	UnclearedBalance pgtype.Int4        `json:"uncleared_balance"`
	LastReconciledAt pgtype.Timestamptz `json:"last_reconciled_at"`
}

func (q *Queries) UpdateAccount(ctx context.Context, arg UpdateAccountParams) (Account, error) {
	row := q.db.QueryRow(ctx, updateAccount,
		arg.ID,
		arg.BudgetID,
		arg.Name,
		arg.Type,
		arg.Closed,
		arg.Note,
		arg.Balance,
		arg.ClearedBalance,
		arg.UnclearedBalance,
		arg.LastReconciledAt,
	)
	var i Account
	err := row.Scan(
		&i.ID,
		&i.BudgetID,
		&i.Name,
		&i.Type,
		&i.Closed,
		&i.Note,
		&i.Balance,
		&i.ClearedBalance,
		&i.UnclearedBalance,
		&i.LastReconciledAt,
	)
	return i, err
}
