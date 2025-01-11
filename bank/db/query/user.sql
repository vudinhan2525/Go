-- name: CreateUser :one
INSERT INTO users (
   hashed_password, full_name, email
) VALUES (
  $1, $2, $3
)
RETURNING *;

-- name: GetUser :one
SELECT * FROM users
WHERE user_id = $1 LIMIT 1;