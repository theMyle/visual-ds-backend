-- name: UpsertSimulatorProgress :exec
-- Store or update a user's progress for a specific simulator path
INSERT INTO simulator_progress (user_id, simulator_category, path, is_completed)
VALUES ($1, $2, $3, $4)
ON CONFLICT (user_id, path) DO UPDATE
SET 
    is_completed = EXCLUDED.is_completed,
    updated_at = now();

-- name: GetSimulatorProgress :one
-- Get a user's progress for a specific simulator
SELECT * FROM simulator_progress
WHERE user_id = $1 AND path = $2;

-- name: ListUserSimulatorProgress :many
-- List all simulator progress related to a user, ordered by category
SELECT * FROM simulator_progress
WHERE user_id = $1
ORDER BY simulator_category, path;

-- name: ListUserSimulatorProgressForCategory :many
-- Get all progress for a specific user and specific category
SELECT * FROM simulator_progress
WHERE user_id = $1 AND simulator_category = $2;