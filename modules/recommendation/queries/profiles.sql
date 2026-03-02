-- name: CreateProfile :one
INSERT INTO profiles(user_id, dance_styles, latitude, longitude, visible, data)
    VALUES (@user_id, @dance_styles, @latitude, @longitude, @visible, @data)
RETURNING *;

-- name: GetProfileByUserID :one
SELECT * FROM profiles WHERE user_id = @user_id;

-- name: UpdateProfile :exec
UPDATE profiles SET
    dance_styles = @dance_styles,
    latitude = @latitude,
    longitude = @longitude,
    visible = @visible,
    data = @data,
    updated_at = CURRENT_TIMESTAMP
WHERE user_id = @user_id;

-- name: UpdateProfileData :exec
UPDATE profiles SET
    data = @data,
    updated_at = CURRENT_TIMESTAMP
WHERE user_id = @user_id;

-- name: GetProfilePreview :one
SELECT
    p.user_id, p.dance_styles, p.data, p.visible,
    u.profile_data
FROM profiles p
JOIN users u ON u.id = p.user_id
WHERE p.user_id = @user_id AND p.visible = true;

-- name: DeleteProfile :exec
DELETE FROM profiles WHERE user_id = @user_id;

-- name: FindNearbyVisibleProfiles :many
SELECT
    p.id, p.user_id, p.dance_styles, p.data,
    p.latitude, p.longitude,
    u.profile_data,
    (6371 * acos(
        cos(radians(@latitude::double precision)) *
        cos(radians(p.latitude)) *
        cos(radians(p.longitude) - radians(@longitude::double precision)) +
        sin(radians(@latitude::double precision)) *
        sin(radians(p.latitude))
    ))::double precision AS distance_km
FROM profiles p
JOIN users u ON u.id = p.user_id
WHERE p.visible = true
  AND p.user_id != @user_id
  AND p.latitude IS NOT NULL
  AND p.longitude IS NOT NULL
  AND NOT (p.user_id = ANY(@exclude_ids::uuid[]))
ORDER BY distance_km ASC
LIMIT @limit_val;
