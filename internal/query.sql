-- name: GetUsers :many
SELECT * FROM users
ORDER BY name;

-- name: GetUserById :one
SELECT * FROM users
WHERE id = $1;

-- name: CreateUser :one
INSERT INTO users (
    id, role_id, name, email, password_hash
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;

-- name: CreateRole :one
INSERT INTO roles (name)
VALUES ($1)
RETURNING *;

-- name: GetRoleByName :one
SELECT id FROM roles
WHERE name = $1;