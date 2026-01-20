-- name: UpsertUserPreferences :one
INSERT INTO user_preferences (
    user_id,
    theme,
    notifications_enabled,
    metadata
) VALUES (
    $1,
    $2,
    $3,
    $4
)
ON CONFLICT (user_id) DO UPDATE SET
    theme = EXCLUDED.theme,
    notifications_enabled = EXCLUDED.notifications_enabled,
    metadata = EXCLUDED.metadata,
    updated_at = NOW()
RETURNING
    id,
    user_id,
    theme,
    notifications_enabled,
    metadata,
    created_at,
    updated_at;

-- name: GetPreferencesByUserID :one
SELECT
    id,
    user_id,
    theme,
    notifications_enabled,
    metadata,
    created_at,
    updated_at
FROM user_preferences
WHERE user_id = $1;