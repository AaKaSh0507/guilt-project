-- name: CreateEntry :one
INSERT INTO guilt_entries (
    session_id,
    entry_text,
    guilt_level
) VALUES (
    $1,
    $2,
    $3
)
RETURNING
    id,
    session_id,
    entry_text,
    guilt_level,
    roast_text,
    created_at,
    updated_at;

-- name: ListEntriesBySession :many
SELECT
    id,
    session_id,
    entry_text,
    guilt_level,
    roast_text,
    created_at,
    updated_at
FROM guilt_entries
WHERE session_id = $1
ORDER BY created_at ASC;

-- name: UpdateRoast :exec
UPDATE guilt_entries SET roast_text = $2 WHERE id = $1;