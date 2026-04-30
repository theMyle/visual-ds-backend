
-- name: GetLessonCategories :many
SELECT lc.*, COUNT(l.lesson_id) as lesson_count
FROM lesson_categories lc
LEFT JOIN lessons l ON lc.category_id = l.category_id
GROUP BY lc.category_id
ORDER BY lc.order_index;

-- name: GetLessonCategoryBySlug :one
SELECT * FROM lesson_categories WHERE slug = $1;

-- name: GetLessonsByCategorySlug :many
SELECT l.* FROM lessons l
JOIN lesson_categories lc ON l.category_id = lc.category_id
WHERE lc.slug = $1
ORDER BY l.order_index;

-- name: GetLessonBySlug :one
SELECT l.* FROM lessons l
JOIN lesson_categories lc ON l.category_id = lc.category_id
WHERE lc.slug = $1 AND l.slug = $2;

-- name: GetLessonByID :one
SELECT * FROM lessons WHERE lesson_id = $1;

-- name: CreateLessonCategory :one
INSERT INTO lesson_categories (slug, title, description, order_index)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: CreateLesson :one
INSERT INTO lessons (category_id, slug, title, content, order_index)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdateLesson :one
UPDATE lessons
SET title = $2, content = $3, order_index = $4, updated_at = now()
WHERE lesson_id = $1
RETURNING *;

-- name: DeleteLesson :exec
DELETE FROM lessons WHERE lesson_id = $1;

-- name: UpdateLessonCategory :one
UPDATE lesson_categories
SET title = $2, slug = $3, description = $4, order_index = $5, updated_at = now()
WHERE category_id = $1
RETURNING *;

-- name: DeleteLessonCategory :exec
DELETE FROM lesson_categories WHERE category_id = $1;
