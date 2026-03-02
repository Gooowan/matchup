-- name: UpsertPreferences :one
INSERT INTO user_preferences(user_id, preferred_styles, preferred_role, min_level, max_level, min_height_cm, max_height_cm, min_age, max_age, max_distance_km, gender_preference)
    VALUES (@user_id, @preferred_styles, @preferred_role, @min_level, @max_level, @min_height_cm, @max_height_cm, @min_age, @max_age, @max_distance_km, @gender_preference)
ON CONFLICT (user_id)
    DO UPDATE SET
        preferred_styles = EXCLUDED.preferred_styles,
        preferred_role = EXCLUDED.preferred_role,
        min_level = EXCLUDED.min_level,
        max_level = EXCLUDED.max_level,
        min_height_cm = EXCLUDED.min_height_cm,
        max_height_cm = EXCLUDED.max_height_cm,
        min_age = EXCLUDED.min_age,
        max_age = EXCLUDED.max_age,
        max_distance_km = EXCLUDED.max_distance_km,
        gender_preference = EXCLUDED.gender_preference,
        updated_at = CURRENT_TIMESTAMP
RETURNING *;

-- name: GetPreferences :one
SELECT * FROM user_preferences WHERE user_id = @user_id;
