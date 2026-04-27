-- name: CreateSimulator :one
INSERT INTO simulators (id, slug, name, description, initial_code, is_active)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetSimulatorBySlug :one
SELECT * FROM simulators
WHERE slug = $1 AND is_active = TRUE;

-- name: ListSimulators :many
SELECT * FROM simulators
WHERE is_active = TRUE;

-- name: UpdateSimulator :one
UPDATE simulators
SET name = $2, slug = $3, description = $4, initial_code = $5, is_active = $6
WHERE id = $1
RETURNING *;

-- name: CreateChallenge :one
INSERT INTO simulator_challenges (
    id, simulator_id, slug, title, description, order_index, 
    initial_code, program_structure, test_cases, capacity, 
    next_challenge_id, is_active
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING *;

-- name: GetChallengeBySlug :one
SELECT 
    c.id, c.simulator_id, c.slug, c.title, c.description, c.order_index, 
    (s.initial_code || COALESCE(c.initial_code, ''))::TEXT as initial_code,
    c.program_structure, c.test_cases, c.capacity, c.next_challenge_id, c.is_active, 
    next_c.slug as next_challenge_slug
FROM simulator_challenges c
JOIN simulators s ON c.simulator_id = s.id
LEFT JOIN simulator_challenges next_c ON c.next_challenge_id = next_c.id
WHERE c.simulator_id = $1 AND c.slug = $2 AND c.is_active = TRUE;

-- name: GetNextChallengeByOrder :one
SELECT slug FROM simulator_challenges
WHERE simulator_id = $1 AND order_index > $2 AND is_active = TRUE
ORDER BY order_index ASC
LIMIT 1;

-- name: ListChallengesForSimulator :many
SELECT * FROM simulator_challenges
WHERE simulator_id = $1 AND is_active = TRUE
ORDER BY order_index ASC;

-- name: GetSimulatorCurriculum :many
SELECT 
    s.id as simulator_id,
    s.slug as simulator_slug,
    s.name as simulator_name,
    s.description as simulator_description,
    s.initial_code as simulator_initial_code,
    s.is_active as simulator_is_active,
    c.id as challenge_id,
    c.slug as challenge_slug,
    c.title as challenge_title,
    c.order_index
FROM simulators s
LEFT JOIN simulator_challenges c ON s.id = c.simulator_id
WHERE s.is_active = TRUE AND (c.is_active = TRUE OR c.is_active IS NULL)
ORDER BY s.name, c.order_index;

-- name: GetSimulatorCurriculumAdmin :many
SELECT 
    s.id as simulator_id,
    s.slug as simulator_slug,
    s.name as simulator_name,
    s.description as simulator_description,
    s.initial_code as simulator_initial_code,
    s.is_active as simulator_is_active,
    c.id as challenge_id,
    c.slug as challenge_slug,
    c.title as challenge_title,
    c.order_index
FROM simulators s
LEFT JOIN simulator_challenges c ON s.id = c.simulator_id
ORDER BY s.name, c.order_index;

-- name: DeleteSimulatorChallenges :exec
DELETE FROM simulator_challenges
WHERE simulator_id = $1;

-- name: GetChallengeByID :one
SELECT * FROM simulator_challenges
WHERE id = $1;

-- name: UpdateChallenge :one
UPDATE simulator_challenges
SET 
    slug = $2, 
    title = $3, 
    description = $4, 
    order_index = $5, 
    initial_code = $6, 
    program_structure = $7, 
    test_cases = $8, 
    capacity = $9, 
    next_challenge_id = $10, 
    is_active = $11
WHERE id = $1
RETURNING *;

-- name: DeleteChallenge :exec
DELETE FROM simulator_challenges
WHERE id = $1;
