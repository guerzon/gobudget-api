-- name: GetCategories :many
SELECT * FROM categories WHERE category_group_id = $1;

-- name: GetCategory :one
SELECT * FROM categories WHERE id = $1;

-- name: CreateCategory :one
INSERT INTO categories (
    category_group_id,
    name
) VALUES (
    $1, $2
)
RETURNING *;

-- name: UpdateCategory :one
UPDATE categories SET name = $1 WHERE id = $2 RETURNING *;

-- name: DeleteCategory :exec
DELETE FROM categories WHERE id = $1;

-- name: DeleteCategories :exec
DELETE FROM categories WHERE category_group_id = $1;
