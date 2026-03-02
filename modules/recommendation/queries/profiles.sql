-- name: CreateProfile :one
INSERT INTO profiles(user_id, dance_styles, dance_role, dance_level, height_cm, bio, birth_date, gender, city, latitude, longitude, visible)
    VALUES (@user_id, @dance_styles, @dance_role, @dance_level, @height_cm, @bio, @birth_date, @gender, @city, @latitude, @longitude, @visible)
RETURNING *;

-- name: GetProfileByUserID :one
SELECT * FROM profiles WHERE user_id = @user_id;

-- name: UpdateProfile :exec
UPDATE profiles SET
    dance_styles = @dance_styles,
    dance_role = @dance_role,
    dance_level = @dance_level,
    height_cm = @height_cm,
    bio = @bio,
    birth_date = @birth_date,
    gender = @gender,
    city = @city,
    latitude = @latitude,
    longitude = @longitude,
    visible = @visible,
    updated_at = CURRENT_TIMESTAMP
WHERE user_id = @user_id;

-- name: UpdateProfileMediaURLs :exec
UPDATE profiles SET
    media_urls = @media_urls,
    updated_at = CURRENT_TIMESTAMP
WHERE user_id = @user_id;

-- name: GetProfilePreview :one
SELECT
    p.user_id, p.dance_styles, p.dance_role, p.dance_level,
    p.height_cm, p.bio, p.gender, p.city, p.media_urls,
    u.profile_data
FROM profiles p
JOIN users u ON u.id = p.user_id
WHERE p.user_id = @user_id AND p.visible = true;

-- name: DeleteProfile :exec
DELETE FROM profiles WHERE user_id = @user_id;

-- name: FindNearbyVisibleProfiles :many
SELECT
    p.id, p.user_id, p.dance_styles, p.dance_role, p.dance_level,
    p.height_cm, p.bio, p.birth_date, p.gender, p.city,
    p.latitude, p.longitude, p.media_urls,
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
