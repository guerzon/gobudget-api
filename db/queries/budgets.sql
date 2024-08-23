-- name: CreateBudget :one
INSERT INTO budgets (
    owner_username,
    name,
    currency_code
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: GetBudgets :many
SELECT * FROM budgets WHERE owner_username = $1;

-- name: GetBudget :one
SELECT * FROM budgets WHERE id = $1 AND owner_username = $2;

-- name: GetBudgetDetails :one
SELECT * FROM budgets WHERE owner_username = $1 AND name = $2 AND currency_code = $3;

-- name: DeleteBudget :exec
DELETE FROM budgets WHERE id = $1;

-- name: DeleteBudgets :exec
DELETE FROM budgets WHERE owner_username = $1;
