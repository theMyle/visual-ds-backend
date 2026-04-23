-- +goose Up
CREATE TABLE assessments (
    id TEXT PRIMARY KEY,
    category TEXT NOT NULL
);

CREATE TABLE questions (
    id TEXT PRIMARY KEY,
    assessment_id TEXT NOT NULL REFERENCES assessments(id) ON DELETE CASCADE,
    text TEXT NOT NULL,
    image_url TEXT,
    type TEXT NOT NULL,
    feedback_correct TEXT NOT NULL,
    feedback_incorrect TEXT NOT NULL
);

CREATE INDEX idx_questions_assessment_id ON questions(assessment_id);

CREATE TABLE choices (
    id TEXT NOT NULL,
    question_id TEXT NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    text TEXT NOT NULL,
    is_correct BOOLEAN NOT NULL DEFAULT FALSE,
    PRIMARY KEY (question_id, id)
);

CREATE INDEX idx_choices_question_id ON choices(question_id);

-- +goose Down
DROP TABLE choices;
DROP TABLE questions;
DROP TABLE assessments;
