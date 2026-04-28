-- +goose Up
CREATE TABLE user_seen_questions (
    user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    assessment_id TEXT NOT NULL REFERENCES assessments(id) ON DELETE CASCADE,
    question_id TEXT NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, assessment_id, question_id)
);

-- +goose Down
DROP TABLE user_seen_questions;
