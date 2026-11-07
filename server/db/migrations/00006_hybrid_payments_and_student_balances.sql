-- +goose Up
ALTER TABLE payments_module.student_payments
    ALTER COLUMN paid_until DROP NOT NULL;

ALTER TABLE payments_module.student_payments
    ADD COLUMN sessions_added BIGINT;

ALTER TABLE payments_module.student_payments
    ADD CONSTRAINT check_student_payments_paid_until_or_sessions_added
        CHECK (paid_until IS NOT NULL OR sessions_added IS NOT NULL);

ALTER TABLE payments_module.student_payments
    ADD CONSTRAINT check_student_payments_sessions_added_positive
        CHECK (sessions_added IS NULL OR sessions_added > 0);

DROP INDEX IF EXISTS payments_module.idx_student_payments_student_paid_until_unique;

CREATE INDEX idx_student_payments_student_paid_until
    ON payments_module.student_payments USING btree (student_id, paid_until DESC NULLS LAST);

CREATE TABLE payments_module.student_balances
(
    owner_id      UUID        NOT NULL REFERENCES users_module.oauth2_users (id) ON DELETE CASCADE,
    student_id    UUID        NOT NULL REFERENCES students_module.students (id) ON DELETE CASCADE,
    paid_until    TIMESTAMPTZ,
    sessions_left BIGINT      NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (owner_id, student_id),
    CONSTRAINT check_student_balances_sessions_left_non_negative CHECK (sessions_left >= 0)
);

CREATE TRIGGER trg_student_balances_updated_at
    BEFORE UPDATE
    ON payments_module.student_balances
    FOR EACH ROW EXECUTE FUNCTION public.set_current_timestamp_updated_at();

INSERT INTO payments_module.student_balances (owner_id, student_id, paid_until, sessions_left)
SELECT s.owner_id,
       s.id,
       latest_payments.max_paid_until,
       0
FROM students_module.students s
         LEFT JOIN (SELECT sp.student_id,
                           t.owner_id,
                           MAX(sp.paid_until) AS max_paid_until
                    FROM payments_module.student_payments sp
                             JOIN payments_module.transactions t ON t.id = sp.transaction_id
                    GROUP BY sp.student_id, t.owner_id) latest_payments
                   ON latest_payments.student_id = s.id AND
                      latest_payments.owner_id =
                      s.owner_id ON CONFLICT (owner_id, student_id) DO NOTHING;

-- +goose Down
DROP TABLE IF EXISTS payments_module.student_balances;

DROP INDEX IF EXISTS payments_module.idx_student_payments_student_paid_until;

CREATE UNIQUE INDEX idx_student_payments_student_paid_until_unique
    ON payments_module.student_payments USING btree (student_id, paid_until);

ALTER TABLE payments_module.student_payments
DROP
CONSTRAINT IF EXISTS check_student_payments_sessions_added_positive,
    DROP
CONSTRAINT IF EXISTS check_student_payments_paid_until_or_sessions_added,
    DROP
COLUMN IF EXISTS sessions_added;

ALTER TABLE payments_module.student_payments
    ALTER COLUMN paid_until SET NOT NULL;
