
-- +goose Up
CREATE TABLE quiz_results (
    user_id UUID NOT NULL,
    quiz_id TEXT NOT NULL,
    score INT NOT NULL,
    total_items INT NOT NULL,
    taken_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    PRIMARY KEY(user_id, quiz_id),

    CONSTRAINT fk_userid
        FOREIGN KEY (user_id) 
        REFERENCES users (user_id)
        ON DELETE CASCADE
);

-- +goose Down
DROP TABLE quiz_results;

