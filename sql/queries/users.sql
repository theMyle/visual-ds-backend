
-- name: CreateUser :one
INSERT INTO users(
    clerk_id,
    course_id,
    first_name,
    middle_name,
    last_name,
    block_id
)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpdateUser :one
UPDATE users
SET
    course_id = COALESCE(sqlc.narg('course_id'), course_id),
    first_name = COALESCE(sqlc.narg('first_name'), first_name),
    middle_name = COALESCE(sqlc.narg('middle_name'), middle_name),
    last_name = COALESCE(sqlc.narg('last_name'), last_name),
    block_id = COALESCE(sqlc.narg('block_id'), block_id),
    updated_at = now()
WHERE clerk_id = sqlc.arg('clerk_id')
RETURNING *;

-- name: GetUserByClearkID :one
SELECT * from users
WHERE clerk_id = $1 LIMIT 1;

-- name: GetUserByID :one
SELECT * from users
WHERE user_id = $1 LIMIT 1;

-- name: GetAllUsers :many
SELECT * from users;

-- name: GetAllUsersByCourse :many
SELECT * from users
WHERE course_id = $1;