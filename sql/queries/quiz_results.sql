
-- name: CreateQuizResultEntry :one
INSERT INTO quiz_results(user_id, quiz_category, quiz_id, score, total_items)
VALUES($1, $2, $3, $4, $5)
RETURNING *;

-- name: DeleteQuizResultEntry :exec
DELETE FROM quiz_results
WHERE
    user_id = $1
    AND quiz_id = $2
RETURNING *;

-- name: DeleteAllQuizResultEntry :exec
DELETE FROM quiz_results
WHERE
    user_id = $1
RETURNING *;

-- name: GetAllQuizResultEntryByUser :many
SELECT * FROM quiz_results
WHERE
    user_id = $1;

-- name: GetAllQuizResultEntryByCategory :many
SELECT * FROM quiz_results
WHERE
    user_id = $1 AND quiz_category = $2;