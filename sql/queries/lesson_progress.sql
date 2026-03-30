
-- name: CreateLessonProgressEntry :one
INSERT INTO lesson_progress(user_id, lesson_slug)
VALUES (
    $1,
    $2
)
RETURNING *;

-- name: DeleteLessonProgress :one
DELETE FROM lesson_progress
WHERE
    user_id = $1
    AND lesson_slug = $2
RETURNING *;

-- name: GetAllLessonProgress :many
SELECT * FROM lesson_progress
WHERE
    user_id = $1
    AND lesson_slug = $2;

-- name: GetLessonProgress :many
SELECT * FROM lesson_progress
WHERE
    user_id = $1;
