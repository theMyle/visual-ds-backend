
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

-- name: GetQuizResultSummariesByUser :many
WITH ranked AS (
    SELECT
        id,
        quiz_category,
        quiz_id,
        score,
        total_items,
        taken_at,
        ROW_NUMBER() OVER (
            PARTITION BY quiz_category, quiz_id
            ORDER BY score DESC, taken_at DESC, id DESC
        ) AS highest_rank,
        ROW_NUMBER() OVER (
            PARTITION BY quiz_category, quiz_id
            ORDER BY taken_at DESC, id DESC
        ) AS most_recent_rank
    FROM quiz_results
    WHERE user_id = $1
),
highest AS (
    SELECT
        quiz_category,
        quiz_id,
        id,
        score,
        total_items,
        taken_at
    FROM ranked
    WHERE highest_rank = 1
),
most_recent AS (
    SELECT
        quiz_category,
        quiz_id,
        id,
        score,
        total_items,
        taken_at
    FROM ranked
    WHERE most_recent_rank = 1
)
SELECT
    h.quiz_category,
    h.quiz_id,
    h.id AS highest_id,
    h.score AS highest_score,
    h.total_items AS highest_total_items,
    h.taken_at AS highest_taken_at,
    mr.id AS most_recent_id,
    mr.score AS most_recent_score,
    mr.total_items AS most_recent_total_items,
    mr.taken_at AS most_recent_taken_at
FROM highest h
JOIN most_recent mr USING (quiz_category, quiz_id)
ORDER BY h.quiz_category, h.quiz_id;