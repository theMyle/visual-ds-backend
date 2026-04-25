-- +goose Up
CREATE TABLE simulators (
    id TEXT PRIMARY KEY,
    slug TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE simulator_challenges (
    id TEXT PRIMARY KEY,
    simulator_id TEXT NOT NULL REFERENCES simulators(id) ON DELETE CASCADE,
    slug TEXT NOT NULL,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    order_index INTEGER NOT NULL,
    initial_code TEXT NOT NULL,
    program_structure JSONB NOT NULL,
    test_cases JSONB NOT NULL,
    capacity JSONB NOT NULL,
    next_challenge_id TEXT REFERENCES simulator_challenges(id) ON DELETE SET NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    UNIQUE(simulator_id, slug)
);

CREATE INDEX idx_challenges_simulator_id ON simulator_challenges(simulator_id);
CREATE INDEX idx_challenges_order ON simulator_challenges(order_index);

-- +goose Down
DROP TABLE simulator_challenges;
DROP TABLE simulators;
