
-- name: GetAccounts :many
SELECT * FROM accounts WHERE budget_id = $1;

-- name: GetAccount :one
SELECT * FROM accounts WHERE budget_id = $1 and id = $2;

-- name: CreateAccount :one
INSERT INTO accounts (
    budget_id,
    name,
    type,
    balance
) VALUES (
    $1, $2, $3, $4
) RETURNING *;

-- name: GetBudgetAccount :one
SELECT * FROM budgets b, accounts a
WHERE b.id = a.budget_id and b.id = $1 and a.id = $2 and b.owner_username = $3;

-- name: UpdateAccount :one
UPDATE accounts
SET
    name = COALESCE(sqlc.narg(name), name),
    type = COALESCE(sqlc.narg(type), type),
    closed = COALESCE(sqlc.narg(closed), closed),
    note = COALESCE(sqlc.narg(note), note),
    balance = COALESCE(sqlc.narg(balance), balance),
    cleared_balance = COALESCE(sqlc.narg(cleared_balance), cleared_balance),
    uncleared_balance = COALESCE(sqlc.narg(uncleared_balance), uncleared_balance),
    last_reconciled_at = COALESCE(sqlc.narg(last_reconciled_at), last_reconciled_at)
WHERE id = $1 AND budget_id = $2
RETURNING *;

-- name: DeleteAccount :exec
DELETE FROM accounts WHERE id = $1;

-- name: DeleteAccounts :exec
DELETE FROM accounts WHERE budget_id = $1;
