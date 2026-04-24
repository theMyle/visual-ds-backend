-- name: SaveQuizResult :one
INSERT INTO quiz_results (
    user_id, quiz_category, quiz_id, score, total_items
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;
