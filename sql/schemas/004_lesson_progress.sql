
-- +goose Up
CREATE TABLE lesson_progress(
    user_id UUID NOT NULL DEFAULT gen_random_uuid(),
    lesson_category TEXT NOT NULL,
    lesson_id TEXT NOT NULL,
    completed_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    PRIMARY KEY(user_id, lesson_category, lesson_id)
);

-- +goose Down
DROP TABLE lesson_progress;