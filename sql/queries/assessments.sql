-- name: CreateAssessment :one
INSERT INTO assessments (
    id, category
) VALUES (
    $1, $2
) RETURNING *;

-- name: CreateQuestion :one
INSERT INTO questions (
    id, assessment_id, text, image_url, type, feedback_correct, feedback_incorrect
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: CreateChoice :one
INSERT INTO choices (
    id, question_id, text, is_correct
) VALUES (
    $1, $2, $3, $4
) RETURNING *;

-- name: BulkCreateAssessments :exec
INSERT INTO assessments (
    id, category
) 
SELECT 
    unnest(@ids::text[]), 
    unnest(@categories::text[]);

-- name: BulkCreateQuestions :exec
INSERT INTO questions (
    id, assessment_id, text, image_url, type, feedback_correct, feedback_incorrect
) 
SELECT 
    unnest(@ids::text[]), 
    unnest(@assessment_ids::text[]), 
    unnest(@texts::text[]), 
    unnest(@image_urls::text[]), 
    unnest(@types::text[]), 
    unnest(@feedbacks_correct::text[]), 
    unnest(@feedbacks_incorrect::text[]);

-- name: BulkCreateChoices :exec
INSERT INTO choices (
    id, question_id, text, is_correct
) 
SELECT 
    unnest(@ids::text[]), 
    unnest(@question_ids::text[]), 
    unnest(@texts::text[]), 
    unnest(@is_corrects::boolean[]);

-- name: GetAssessment :one
SELECT * FROM assessments WHERE category = $1 AND id = $2 LIMIT 1;

-- name: GetAssessmentById :one
SELECT * FROM assessments WHERE id = $1 LIMIT 1;

-- name: GetQuestionsByAssessmentId :many
SELECT * FROM questions WHERE assessment_id = $1 ORDER BY id ASC;

-- name: GetChoicesByQuestionIds :many
SELECT * FROM choices WHERE question_id = ANY(@question_ids::text[]) ORDER BY question_id, id ASC;

-- name: ListAssessments :many
SELECT * FROM assessments ORDER BY category ASC, id ASC LIMIT $1;

-- name: DeleteAssessment :exec
DELETE FROM assessments WHERE id = $1;

-- name: UpdateAssessment :one
UPDATE assessments SET category = $2 WHERE id = $1 RETURNING *;

-- name: DeleteQuestion :exec
DELETE FROM questions WHERE id = $1;
