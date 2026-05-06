-- +goose Up
ALTER TABLE simulator_progress ADD COLUMN last_submitted_code TEXT NOT NULL DEFAULT '';

CREATE TABLE simulator_submissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users (user_id) ON DELETE CASCADE,
    simulator_id TEXT NOT NULL,
    challenge_id TEXT NOT NULL REFERENCES simulator_challenges (id) ON DELETE CASCADE,
    code TEXT NOT NULL,
    status TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_submissions_user_challenge ON simulator_submissions (user_id, challenge_id);

-- +goose Down
DROP TABLE simulator_submissions;
ALTER TABLE simulator_progress DROP COLUMN last_submitted_code;
