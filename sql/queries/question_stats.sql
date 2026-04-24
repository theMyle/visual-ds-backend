-- name: UpdateQuestionStats :exec
INSERT INTO question_stats (question_id, correct, mistakes)
VALUES ($1, $2, $3)
ON CONFLICT (question_id) DO UPDATE SET
    correct = question_stats.correct + EXCLUDED.correct,
    mistakes = question_stats.mistakes + EXCLUDED.mistakes;

-- name: ListQuestionStats :many
SELECT * FROM question_stats;

-- name: GetQuestionStats :one
SELECT * FROM question_stats WHERE question_id = $1;
