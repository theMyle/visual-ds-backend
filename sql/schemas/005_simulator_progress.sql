-- +goose Up
CREATE TABLE simulator_progress (
    user_id UUID NOT NULL,
    simulator_category TEXT NOT NULL, 
    path TEXT NOT NULL, 
    is_completed BOOLEAN NOT NULL DEFAULT FALSE,
    updated_at TIMESTAMPTZ DEFAULT now(),

    CONSTRAINT fk_user
        FOREIGN KEY (user_id)
        REFERENCES users (user_id)
        ON DELETE CASCADE,

    -- Ensures a user has a unique record for every specific path
    PRIMARY KEY (user_id, path)
);

-- Crucial for filtering a user's progress by a specific category
CREATE INDEX idx_user_category_lookup ON simulator_progress (user_id, simulator_category);

-- +goose Down
DROP TABLE simulator_progress;