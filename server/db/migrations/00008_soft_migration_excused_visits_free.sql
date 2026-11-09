-- +goose Up
UPDATE visits_module.visits
SET type = 'FREE'
WHERE status = 'EXCUSED' AND type != 'FREE';

ALTER TABLE visits_module.visits
    ADD CONSTRAINT check_excused_visits_free CHECK (status != 'EXCUSED' OR type = 'FREE');

-- +goose Down
ALTER TABLE visits_module.visits
    DROP CONSTRAINT IF EXISTS check_excused_visits_free;
