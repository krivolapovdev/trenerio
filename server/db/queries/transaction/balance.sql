-- name: FindStudentBalanceByStudentIDAndOwnerID :one
SELECT
    s.id AS student_id,
    (
        SELECT sp.paid_from
        FROM finance_module.student_payments sp
        JOIN finance_module.transactions t ON t.id = sp.transaction_id
        WHERE sp.student_id = s.id AND t.owner_id = s.owner_id AND sp.paid_until IS NOT NULL
        ORDER BY sp.paid_until DESC NULLS LAST
        LIMIT 1
    ) AS paid_from,
    (
        SELECT sp.paid_until
        FROM finance_module.student_payments sp
        JOIN finance_module.transactions t ON t.id = sp.transaction_id
        WHERE sp.student_id = s.id AND t.owner_id = s.owner_id AND sp.paid_until IS NOT NULL
        ORDER BY sp.paid_until DESC NULLS LAST
        LIMIT 1
    ) AS paid_until,
    GREATEST(0,
        COALESCE((
            SELECT SUM(sp.sessions_added)
            FROM finance_module.student_payments sp
            JOIN finance_module.transactions t ON t.id = sp.transaction_id
            WHERE sp.student_id = s.id AND t.owner_id = s.owner_id
        ), 0)
        -
        CASE
            WHEN s.free THEN 0
            ELSE COALESCE((
                SELECT COUNT(*)
                FROM visits_module.visits v
                JOIN lessons_module.lessons l ON l.id = v.lesson_id
                WHERE v.student_id = s.id
                  AND v.owner_id = s.owner_id
                  AND v.type = 'REGULAR'
                  AND v.status IN ('PRESENT', 'ABSENT')
                  AND NOT EXISTS (
                      SELECT 1
                      FROM finance_module.student_payments sp
                      JOIN finance_module.transactions t ON t.id = sp.transaction_id
                      WHERE sp.student_id = s.id
                        AND t.owner_id = s.owner_id
                        AND sp.paid_until IS NOT NULL
                        AND (sp.paid_from IS NULL OR date_trunc('day', l.date) >= date_trunc('day', sp.paid_from))
                        AND date_trunc('day', l.date) <= date_trunc('day', sp.paid_until)
                  )
            ), 0)
        END
    )::bigint AS sessions_left,
    COALESCE((
        SELECT SUM(sp.sessions_added)
        FROM finance_module.student_payments sp
        JOIN finance_module.transactions t ON t.id = sp.transaction_id
        WHERE sp.student_id = s.id AND t.owner_id = s.owner_id
    ), 0)::bigint AS total_sessions_added
FROM students_module.students s
WHERE s.owner_id = $1 AND s.id = $2;

-- name: FindStudentBalancesByOwnerID :many
SELECT
    s.id AS student_id,
    (
        SELECT sp.paid_from
        FROM finance_module.student_payments sp
        JOIN finance_module.transactions t ON t.id = sp.transaction_id
        WHERE sp.student_id = s.id AND t.owner_id = s.owner_id AND sp.paid_until IS NOT NULL
        ORDER BY sp.paid_until DESC NULLS LAST
        LIMIT 1
    ) AS paid_from,
    (
        SELECT sp.paid_until
        FROM finance_module.student_payments sp
        JOIN finance_module.transactions t ON t.id = sp.transaction_id
        WHERE sp.student_id = s.id AND t.owner_id = s.owner_id AND sp.paid_until IS NOT NULL
        ORDER BY sp.paid_until DESC NULLS LAST
        LIMIT 1
    ) AS paid_until,
    GREATEST(0,
        COALESCE((
            SELECT SUM(sp.sessions_added)
            FROM finance_module.student_payments sp
            JOIN finance_module.transactions t ON t.id = sp.transaction_id
            WHERE sp.student_id = s.id AND t.owner_id = s.owner_id
        ), 0)
        -
        CASE
            WHEN s.free THEN 0
            ELSE COALESCE((
                SELECT COUNT(*)
                FROM visits_module.visits v
                JOIN lessons_module.lessons l ON l.id = v.lesson_id
                WHERE v.student_id = s.id
                  AND v.owner_id = s.owner_id
                  AND v.type = 'REGULAR'
                  AND v.status IN ('PRESENT', 'ABSENT')
                  AND NOT EXISTS (
                      SELECT 1
                      FROM finance_module.student_payments sp
                      JOIN finance_module.transactions t ON t.id = sp.transaction_id
                      WHERE sp.student_id = s.id
                        AND t.owner_id = s.owner_id
                        AND sp.paid_until IS NOT NULL
                        AND (sp.paid_from IS NULL OR date_trunc('day', l.date) >= date_trunc('day', sp.paid_from))
                        AND date_trunc('day', l.date) <= date_trunc('day', sp.paid_until)
                  )
            ), 0)
        END
    )::bigint AS sessions_left,
    COALESCE((
        SELECT SUM(sp.sessions_added)
        FROM finance_module.student_payments sp
        JOIN finance_module.transactions t ON t.id = sp.transaction_id
        WHERE sp.student_id = s.id AND t.owner_id = s.owner_id
    ), 0)::bigint AS total_sessions_added
FROM students_module.students s
WHERE s.owner_id = $1;

-- name: FindStudentBalancesByStudentIDsAndOwnerID :many
SELECT
    s.id AS student_id,
    (
        SELECT sp.paid_from
        FROM finance_module.student_payments sp
        JOIN finance_module.transactions t ON t.id = sp.transaction_id
        WHERE sp.student_id = s.id AND t.owner_id = s.owner_id AND sp.paid_until IS NOT NULL
        ORDER BY sp.paid_until DESC NULLS LAST
        LIMIT 1
    ) AS paid_from,
    (
        SELECT sp.paid_until
        FROM finance_module.student_payments sp
        JOIN finance_module.transactions t ON t.id = sp.transaction_id
        WHERE sp.student_id = s.id AND t.owner_id = s.owner_id AND sp.paid_until IS NOT NULL
        ORDER BY sp.paid_until DESC NULLS LAST
        LIMIT 1
    ) AS paid_until,
    GREATEST(0,
        COALESCE((
            SELECT SUM(sp.sessions_added)
            FROM finance_module.student_payments sp
            JOIN finance_module.transactions t ON t.id = sp.transaction_id
            WHERE sp.student_id = s.id AND t.owner_id = s.owner_id
        ), 0)
        -
        CASE
            WHEN s.free THEN 0
            ELSE COALESCE((
                SELECT COUNT(*)
                FROM visits_module.visits v
                JOIN lessons_module.lessons l ON l.id = v.lesson_id
                WHERE v.student_id = s.id
                  AND v.owner_id = s.owner_id
                  AND v.type = 'REGULAR'
                  AND v.status IN ('PRESENT', 'ABSENT')
                  AND NOT EXISTS (
                      SELECT 1
                      FROM finance_module.student_payments sp
                      JOIN finance_module.transactions t ON t.id = sp.transaction_id
                      WHERE sp.student_id = s.id
                        AND t.owner_id = s.owner_id
                        AND sp.paid_until IS NOT NULL
                        AND (sp.paid_from IS NULL OR date_trunc('day', l.date) >= date_trunc('day', sp.paid_from))
                        AND date_trunc('day', l.date) <= date_trunc('day', sp.paid_until)
                  )
            ), 0)
        END
    )::bigint AS sessions_left,
    COALESCE((
        SELECT SUM(sp.sessions_added)
        FROM finance_module.student_payments sp
        JOIN finance_module.transactions t ON t.id = sp.transaction_id
        WHERE sp.student_id = s.id AND t.owner_id = s.owner_id
    ), 0)::bigint AS total_sessions_added
FROM students_module.students s
WHERE s.owner_id = $1 AND s.id = ANY($2::uuid[]);
