
-- name: CreateLessonProgressEntry :one
INSERT INTO lesson_progress(user_id, lesson_slug)
VALUES (
    (SELECT user_id FROM users WHERE clerk_id = $1),
    $2
)
RETURNING *;

-- name: DeleteLessonProgress :one
DELETE FROM lesson_progress
WHERE
    user_id = (SELECT user_id FROM users WHERE clerk_id = $1)
    AND lesson_slug = $2
RETURNING *;