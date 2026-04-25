-- +goose Up
ALTER TABLE simulators ADD COLUMN initial_code TEXT NOT NULL DEFAULT '';
ALTER TABLE simulator_challenges ALTER COLUMN initial_code DROP NOT NULL;
ALTER TABLE simulator_challenges ALTER COLUMN initial_code SET DEFAULT NULL;

-- +goose Down
UPDATE simulator_challenges SET initial_code = '' WHERE initial_code IS NULL;
ALTER TABLE simulator_challenges ALTER COLUMN initial_code SET NOT NULL;
ALTER TABLE simulators DROP COLUMN initial_code;
