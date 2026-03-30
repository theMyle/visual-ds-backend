-- name: CreateLessonProgressEntry :one
INSERT INTO lesson_progress(user_id, lesson_category, lesson_id)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, lesson_category, lesson_id) 
DO UPDATE SET completed_at = now()
RETURNING *;

-- name: DeleteLessonProgress :one
DELETE FROM lesson_progress
WHERE user_id = $1 
  AND lesson_category = $2
  AND lesson_id = $3
RETURNING *;

-- name: GetAllLessonProgressByCategory :many
SELECT * FROM lesson_progress
WHERE user_id = $1 
  AND lesson_category = $2;

-- name: GetAllLessonProgressByUser :many
SELECT * FROM lesson_progress
WHERE user_id = $1;

-- name: GetLessonProgressByID :one
SELECT * FROM lesson_progress
WHERE user_id = $1 
  AND lesson_category = $2
  AND lesson_id = $3;