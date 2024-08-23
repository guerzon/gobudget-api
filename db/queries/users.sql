-- name: GetUsers :many
SELECT * FROM users;

-- name: GetUserById :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = $1;

-- name: CreateUser :one
INSERT INTO users (
    username,
    password,
    email,
    email_verified,
    last_password_change
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: UpdateUser :one
UPDATE users
SET
    password = COALESCE(sqlc.narg(password), password),
    email = COALESCE(sqlc.narg(email), email),
    email_verified = COALESCE(sqlc.narg(email_verified), email_verified),
    last_password_change = COALESCE(sqlc.narg(last_password_change), last_password_change)
WHERE username = sqlc.arg(username)
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users WHERE username = $1;
