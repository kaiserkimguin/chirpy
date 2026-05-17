-- name: CreateUser :one
INSERT INTO users (
  id, created_at, updated_at, email, hashed_password
) VALUES ( 
  gen_random_uuid(), NOW(), NOW(), $1, $2
) RETURNING *;

-- name: GetUser :one
SELECT *
FROM users
WHERE email = $1;

-- name: GetUserByID :one
SELECT *
FROM users
WHERE id = $1;

-- name: DeleteUsers :many
DELETE FROM users
RETURNING *;

-- name: UpdateUserData :one
UPDATE users
SET email = $1, hashed_password = $2, updated_at = NOW()
WHERE id = $3
RETURNING *;

-- name: UpdateUserStatus :one
UPDATE users
SET is_red = true, updated_at = NOW()
WHERE id = $1
RETURNING *;
