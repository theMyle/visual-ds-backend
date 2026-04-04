
-- +goose Up
CREATE TABLE quiz_results (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    quiz_category TEXT NOT NULL,
    quiz_id TEXT NOT NULL,
    score INT NOT NULL,
    total_items INT NOT NULL,
    taken_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT fk_userid
        FOREIGN KEY (user_id) 
        REFERENCES users (user_id)
        ON DELETE CASCADE
);

CREATE INDEX idx_quiz_results_user_category ON quiz_results (user_id, quiz_category, taken_at DESC);

-- +goose Down
DROP TABLE quiz_results;

