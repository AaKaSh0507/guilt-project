-- name: CreateScore :one
INSERT INTO guilt_scores (
    session_id,
    entry_id,
    aggregate_score,
    meta
) VALUES (
    $1,
    $2,
    $3,
    $4
)
RETURNING
    id,
    session_id,
    entry_id,
    aggregate_score,
    meta,
    created_at,
    updated_at;

-- name: GetScoreBySession :one
SELECT
    id,
    session_id,
    entry_id,
    aggregate_score,
    meta,
    created_at,
    updated_at
FROM guilt_scores
WHERE session_id = $1;

-- name: GetScoreByEntry :one
SELECT
    id,
    session_id,
    entry_id,
    aggregate_score,
    meta,
    created_at,
    updated_at
FROM guilt_scores
WHERE entry_id = $1;