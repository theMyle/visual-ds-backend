-- +goose Up
ALTER TABLE assessments ADD COLUMN max_attempts INT DEFAULT NULL;

-- +goose Down
ALTER TABLE assessments DROP COLUMN max_attempts;
