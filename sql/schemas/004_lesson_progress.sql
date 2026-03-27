
-- +goose Up
CREATE TABLE lesson_progress(
    user_id UUID NOT NULL DEFAULT gen_random_uuid(),
    lesson_slug TEXT NOT NULL,
    completed_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    PRIMARY KEY(user_id, lesson_slug)
);

-- +goose Down
DROP TABLE lesson_progress;