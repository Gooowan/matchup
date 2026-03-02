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
