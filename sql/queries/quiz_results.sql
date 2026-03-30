
-- name: CreateQuizResultEntry :one
INSERT INTO quiz_results(user_id, quiz_id, score, total_items)
VALUES($1, $2, $3, $4)
RETURNING *;

-- name: DeleteQuizResultEntry :one
DELETE FROM quiz_results
WHERE
    user_id = $1
    AND quiz_id = $2
RETURNING *;

-- name: DeleteAllQuizResultEntry :many
DELETE FROM quiz_results
WHERE
    user_id = $1
RETURNING *;

-- name: GetAllQuizResultEntry :many
SELECT * FROM quiz_results
WHERE
    user_id = $1;