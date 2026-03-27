
-- +goose Up
CREATE TABLE courses (
    course_id UUID PRIMARY KEY,
    course_name TEXT NOT NULL
);

-- +goose Down
DROP TABLE courses;