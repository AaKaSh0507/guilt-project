-- name: CreateSession :one
INSERT INTO guilt_sessions (
    user_id,
    notes
) VALUES (
    $1,
    $2
)
RETURNING
    id,
    user_id,
    start_time,
    end_time,
    notes,
    created_at,
    updated_at;

-- name: EndSession :one
UPDATE guilt_sessions
SET
    end_time = NOW(),
    updated_at = NOW()
WHERE id = $1
RETURNING
    id,
    user_id,
    start_time,
    end_time,
    notes,
    created_at,
    updated_at;

-- name: GetSessionByID :one
SELECT
    id,
    user_id,
    start_time,
    end_time,
    notes,
    created_at,
    updated_at
FROM guilt_sessions
WHERE id = $1;

-- name: ListSessionsByUser :many
SELECT
    id,
    user_id,
    start_time,
    end_time,
    notes,
    created_at,
    updated_at
FROM guilt_sessions
WHERE user_id = $1
ORDER BY start_time DESC
LIMIT $2 OFFSET $3;