-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, name, email)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetUserById :one
SELECT * FROM users
WHERE id = $1;

-- name: GetUserIDByEmail :one
SELECT id FROM users
WHERE email = $1;
