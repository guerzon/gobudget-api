-- name: CreateVerifyEmails :one
INSERT INTO verify_emails (
    username,
    email,
    code,
    expires_at
) VALUES (
    $1, $2, $3, $4
) RETURNING *;

-- name: GetVerifyEmails :one
SELECT * FROM verify_emails WHERE id = $1 AND code = $2;

-- name: GetPendingVerifyEmails :many
SELECT * FROM verify_emails WHERE username = $1 AND used = $2 AND expires_at >= $3;

-- name: UpdateCodeUsed :one
UPDATE verify_emails
SET used = true
WHERE code = $1
RETURNING *;

-- name: DeleteVerifyEmails :exec
DELETE FROM verify_emails WHERE username = $1;
