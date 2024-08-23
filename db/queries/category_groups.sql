-- name: GetCategoryGroupsByBudgetId :many
SELECT * FROM category_groups WHERE budget_id = $1;

-- name: GetCategoryGroup :one
SELECT * FROM category_groups WHERE id = $1;

-- name: CreateCategoryGroup :one
INSERT INTO category_groups (
    budget_id,
    name
) VALUES (
    $1, $2
)
RETURNING *;

-- name: UpdateCategoryGroup :one
UPDATE category_groups SET name = $1 WHERE id = $2 RETURNING *;

-- name: DeleteCategoryGroup :exec
DELETE FROM category_groups WHERE id = $1;

-- name: DeleteCategoryGroups :exec
DELETE FROM category_groups WHERE budget_id = $1;
