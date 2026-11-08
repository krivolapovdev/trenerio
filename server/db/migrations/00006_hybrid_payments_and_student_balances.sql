-- +goose Up
DROP TABLE IF EXISTS payments_module.student_balances;

ALTER TABLE payments_module.student_payments
    ALTER COLUMN paid_until DROP NOT NULL;

ALTER TABLE payments_module.student_payments
    ADD COLUMN paid_from TIMESTAMPTZ;

ALTER TABLE payments_module.student_payments
    ADD COLUMN sessions_added BIGINT;

-- Soft migration: set paid_from from previous payment's paid_until, or NULL if first payment
WITH prev_payments AS (
    SELECT
        sp.transaction_id,
        LAG(sp.paid_until) OVER (
            PARTITION BY sp.student_id
            ORDER BY sp.paid_until ASC, sp.transaction_id ASC
        ) AS prev_paid_until
    FROM payments_module.student_payments sp
)
UPDATE payments_module.student_payments sp
SET paid_from = pp.prev_paid_until
FROM prev_payments pp
WHERE sp.transaction_id = pp.transaction_id;

ALTER TABLE payments_module.student_payments
    ADD CONSTRAINT check_student_payments_paid_until_or_sessions_added
        CHECK (paid_until IS NOT NULL OR sessions_added IS NOT NULL);

ALTER TABLE payments_module.student_payments
    ADD CONSTRAINT check_student_payments_sessions_added_positive
        CHECK (sessions_added IS NULL OR sessions_added > 0);

ALTER TABLE payments_module.student_payments
    ADD CONSTRAINT check_student_payments_paid_range
        CHECK (paid_from IS NULL OR paid_until IS NULL OR paid_from <= paid_until);

DROP INDEX IF EXISTS payments_module.idx_student_payments_student_paid_until_unique;

CREATE INDEX idx_student_payments_student_paid_until
    ON payments_module.student_payments USING btree (student_id, paid_until DESC NULLS LAST);

CREATE INDEX idx_student_payments_student_paid_from
    ON payments_module.student_payments USING btree (student_id, paid_from NULLS FIRST);

-- +goose Down
DROP TABLE IF EXISTS payments_module.student_balances;

DROP INDEX IF EXISTS payments_module.idx_student_payments_student_paid_from;
DROP INDEX IF EXISTS payments_module.idx_student_payments_student_paid_until;

CREATE UNIQUE INDEX idx_student_payments_student_paid_until_unique
    ON payments_module.student_payments USING btree (student_id, paid_until);

ALTER TABLE payments_module.student_payments
    DROP CONSTRAINT IF EXISTS check_student_payments_paid_range,
    DROP CONSTRAINT IF EXISTS check_student_payments_sessions_added_positive,
    DROP CONSTRAINT IF EXISTS check_student_payments_paid_until_or_sessions_added,
    DROP COLUMN IF EXISTS sessions_added,
    DROP COLUMN IF EXISTS paid_from;

ALTER TABLE payments_module.student_payments
    ALTER COLUMN paid_until SET NOT NULL;
