-- name: CreateSimulatorSubmission :one
INSERT INTO simulator_submissions (user_id, simulator_id, challenge_id, code, status)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListUserSubmissionsForChallenge :many
SELECT * FROM simulator_submissions
WHERE user_id = $1 AND challenge_id = $2
ORDER BY created_at DESC;

-- name: DeleteUserSubmissionsForChallenge :exec
DELETE FROM simulator_submissions
WHERE user_id = $1 AND challenge_id = $2;
