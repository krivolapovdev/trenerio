-- +goose Up
CREATE SCHEMA IF NOT EXISTS finance_module;

ALTER TABLE payments_module.transactions SET SCHEMA finance_module;
ALTER TABLE payments_module.student_payments SET SCHEMA finance_module;
ALTER TYPE payments_module.transaction_type SET SCHEMA finance_module;

DROP SCHEMA IF EXISTS payments_module CASCADE;

-- +goose Down
CREATE SCHEMA IF NOT EXISTS payments_module;

ALTER TABLE finance_module.transactions SET SCHEMA payments_module;
ALTER TABLE finance_module.student_payments SET SCHEMA payments_module;
ALTER TYPE finance_module.transaction_type SET SCHEMA payments_module;

DROP SCHEMA IF EXISTS finance_module CASCADE;
