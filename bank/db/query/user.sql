-- name: CreateUser :one
INSERT INTO users (
   hashed_password, full_name, email, role
) VALUES (
  $1, $2, $3, $4
)
RETURNING *;

-- name: GetUser :one
SELECT * FROM users
WHERE user_id = $1 LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1 LIMIT 1;

-- name: UpdateUser :one
UPDATE users
  set email = coalesce(sqlc.narg('email'), email),
  full_name = coalesce(sqlc.narg('full_name'), full_name),
  hashed_password = coalesce(sqlc.narg('hashed_password'), hashed_password),
  password_changed_at = coalesce(sqlc.narg('password_changed_at'), password_changed_at)
WHERE user_id = sqlc.arg('user_id')
RETURNING *;