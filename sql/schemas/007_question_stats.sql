-- +goose Up
CREATE TABLE question_stats (
    question_id TEXT PRIMARY KEY REFERENCES questions(id) ON DELETE CASCADE,
    correct INT NOT NULL DEFAULT 0,
    mistakes INT NOT NULL DEFAULT 0
);

-- +goose Down
DROP TABLE question_stats;
