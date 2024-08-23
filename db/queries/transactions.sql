
-- name: GetTransactions :many
select trans.*
from transactions trans, accounts accts
where trans.account_id = accts.id AND accts.budget_id = $1;

-- name: GetTransactionsView :many
SELECT * FROM transactions_view WHERE budget_id = $1;

-- name: GetTransactionsById :one
SELECT * FROM transactions WHERE id = $1;

-- name: GetTransactionsViewById :one
SELECT * FROM transactions_view WHERE id = $1;

-- name: CreateTransaction :one
INSERT INTO transactions (
    account_id,
    date,
    payee_id,
    category_id,
    memo,
    amount,
    cleared,
    reconciled
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: UpdateTransaction :one
UPDATE transactions
SET
    account_id = COALESCE(sqlc.narg(account_id), account_id),
    date= COALESCE(sqlc.narg(date), date),
    payee_id = COALESCE(sqlc.narg(payee_id), payee_id),
    category_id = COALESCE(sqlc.narg(category_id), category_id),
    memo = COALESCE(sqlc.narg(memo), memo),
    amount = COALESCE(sqlc.narg(amount), amount),
    approved = COALESCE(sqlc.narg(approved), approved),
    cleared = COALESCE(sqlc.narg(cleared), cleared),
    reconciled = COALESCE(sqlc.narg(reconciled), reconciled)
WHERE id = $1
RETURNING *;

-- name: DeleteTransaction :exec
DELETE FROM transactions WHERE id = $1;
