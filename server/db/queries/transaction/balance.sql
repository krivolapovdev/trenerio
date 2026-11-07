-- name: CreateStudentBalance :exec
INSERT INTO payments_module.student_balances (owner_id, student_id, paid_until, sessions_left)
VALUES ($1, $2, NULL, 0)
ON CONFLICT (owner_id, student_id) DO NOTHING;

-- name: FindStudentBalanceByStudentIDAndOwnerID :one
SELECT b.owner_id,
       b.student_id,
       b.paid_until,
       b.sessions_left,
       COALESCE(
           (SELECT SUM(sp.sessions_added)
            FROM payments_module.student_payments sp
            JOIN payments_module.transactions t ON t.id = sp.transaction_id
            WHERE sp.student_id = b.student_id AND t.owner_id = b.owner_id),
           0
       )::bigint AS total_sessions_added,
       b.created_at,
       b.updated_at
FROM payments_module.student_balances b
WHERE b.owner_id = $1 AND b.student_id = $2;

-- name: FindStudentBalanceForUpdate :one
SELECT owner_id, student_id, paid_until, sessions_left, created_at, updated_at
FROM payments_module.student_balances
WHERE owner_id = $1 AND student_id = $2
FOR UPDATE;

-- name: FindStudentBalancesByOwnerID :many
SELECT b.owner_id,
       b.student_id,
       b.paid_until,
       b.sessions_left,
       COALESCE(
           (SELECT SUM(sp.sessions_added)
            FROM payments_module.student_payments sp
            JOIN payments_module.transactions t ON t.id = sp.transaction_id
            WHERE sp.student_id = b.student_id AND t.owner_id = b.owner_id),
           0
       )::bigint AS total_sessions_added,
       b.created_at,
       b.updated_at
FROM payments_module.student_balances b
WHERE b.owner_id = $1;

-- name: FindStudentBalancesByStudentIDsAndOwnerID :many
SELECT b.owner_id,
       b.student_id,
       b.paid_until,
       b.sessions_left,
       COALESCE(
           (SELECT SUM(sp.sessions_added)
            FROM payments_module.student_payments sp
            JOIN payments_module.transactions t ON t.id = sp.transaction_id
            WHERE sp.student_id = b.student_id AND t.owner_id = b.owner_id),
           0
       )::bigint AS total_sessions_added,
       b.created_at,
       b.updated_at
FROM payments_module.student_balances b
WHERE b.owner_id = $1 AND b.student_id = ANY($2::uuid[]);

-- name: ApplyPaymentToStudentBalance :one
INSERT INTO payments_module.student_balances (owner_id, student_id, paid_until, sessions_left)
VALUES ($1, $2, sqlc.narg('paid_until')::timestamptz, COALESCE(sqlc.narg('sessions_added')::bigint, 0))
ON CONFLICT (owner_id, student_id) DO UPDATE
SET paid_until = CASE
        WHEN EXCLUDED.paid_until IS NULL THEN payments_module.student_balances.paid_until
        WHEN payments_module.student_balances.paid_until IS NULL THEN EXCLUDED.paid_until
        ELSE GREATEST(payments_module.student_balances.paid_until, EXCLUDED.paid_until)
    END,
    sessions_left = payments_module.student_balances.sessions_left + EXCLUDED.sessions_left
RETURNING owner_id, student_id, paid_until, sessions_left, created_at, updated_at;

-- name: UpsertStudentBalance :one
INSERT INTO payments_module.student_balances (owner_id, student_id, paid_until, sessions_left)
VALUES ($1, $2, $3, $4)
ON CONFLICT (owner_id, student_id) DO UPDATE
SET paid_until = EXCLUDED.paid_until,
    sessions_left = EXCLUDED.sessions_left
RETURNING owner_id, student_id, paid_until, sessions_left, created_at, updated_at;
