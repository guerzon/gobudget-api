-- name: GetPayees :many
SELECT * FROM payees WHERE budget_id = $1;

-- name: GetPayeeById :one
SELECT * FROM payees WHERE id = $1;

-- name: CreatePayee :one
INSERT INTO payees (
    budget_id,
    name
) VALUES (
    $1, $2
) RETURNING *;

-- name: UpdatePayee :one
UPDATE payees SET name = $1 WHERE budget_id = $2 AND id = $3 RETURNING *;

-- name: DeletePayee :exec
DELETE FROM payees WHERE budget_id = $1 AND id = $2;
