-- name: UpsertUserLocation :one
INSERT INTO user_locations(user_id, latitude, longitude, updated_at)
    VALUES (@user_id, @latitude, @longitude, CURRENT_TIMESTAMP)
ON CONFLICT (user_id)
    DO UPDATE SET
        latitude = EXCLUDED.latitude,
        longitude = EXCLUDED.longitude,
        updated_at = CURRENT_TIMESTAMP
RETURNING *;

-- name: GetUserLocation :one
SELECT * FROM user_locations WHERE user_id = @user_id;

-- name: DeleteUserLocation :exec
DELETE FROM user_locations WHERE user_id = @user_id;

-- name: FindNearbyUsersByCount :many
SELECT
    ul.user_id,
    ul.latitude,
    ul.longitude,
    ul.updated_at,
    (6371 * acos(
        cos(radians(@latitude::double precision)) *
        cos(radians(ul.latitude)) *
        cos(radians(ul.longitude) - radians(@longitude::double precision)) +
        sin(radians(@latitude::double precision)) *
        sin(radians(ul.latitude))
    ))::double precision AS distance_km
FROM user_locations ul
WHERE ul.user_id != @user_id
ORDER BY distance_km ASC
LIMIT @max_results;

-- name: FindNearbyUsersWithinRadius :many
SELECT user_id, latitude, longitude, updated_at, distance_km FROM (
    SELECT
        ul.user_id,
        ul.latitude,
        ul.longitude,
        ul.updated_at,
        (6371 * acos(
            cos(radians(@latitude::double precision)) *
            cos(radians(ul.latitude)) *
            cos(radians(ul.longitude) - radians(@longitude::double precision)) +
            sin(radians(@latitude::double precision)) *
            sin(radians(ul.latitude))
        ))::double precision AS distance_km
    FROM user_locations ul
    WHERE ul.user_id != @user_id
) sub
WHERE sub.distance_km <= @radius_km::double precision
ORDER BY sub.distance_km ASC;

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
