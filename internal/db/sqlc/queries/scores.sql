-- name: CreateScore :one
INSERT INTO guilt_scores (
    session_id,
    aggregate_score,
    meta
) VALUES (
    $1,
    $2,
    $3
)
RETURNING
    id,
    session_id,
    aggregate_score,
    meta,
    created_at,
    updated_at;

-- name: GetScoreBySession :one
SELECT
    id,
    session_id,
    aggregate_score,
    meta,
    created_at,
    updated_at
FROM guilt_scores
WHERE session_id = $1;