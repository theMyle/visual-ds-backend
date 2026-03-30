
-- +goose Up
CREATE TABLE users (
    user_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    clerk_id TEXT UNIQUE NOT NULL,
    course_id UUID,
    first_name VARCHAR(255) NOT NULL,
    middle_name VARCHAR(255),
    last_name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    block_id TEXT,

    CONSTRAINT fk_courseid
        FOREIGN KEY (course_id)
        REFERENCES courses (course_id)
        ON DELETE SET NULL
);

-- +goose Down
DROP TABLE users;