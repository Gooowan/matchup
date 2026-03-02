-- name: UpsertPreferences :one
INSERT INTO user_preferences(user_id, data)
    VALUES (@user_id, @data)
ON CONFLICT (user_id)
    DO UPDATE SET
        data = EXCLUDED.data,
        updated_at = CURRENT_TIMESTAMP
RETURNING *;

-- name: GetPreferences :one
SELECT * FROM user_preferences WHERE user_id = @user_id;
