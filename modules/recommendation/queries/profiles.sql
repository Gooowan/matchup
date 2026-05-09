-- name: CreateProfile :one
INSERT INTO profiles(
    user_id, dance_styles, latitude, longitude, visible,
    gender, birth_date, height_cm, goal, program, categories,
    country, city, ready_to_relocate, ready_to_finance,
    primary_club_id, metadata, data
)
VALUES (
    @user_id, @dance_styles, @latitude, @longitude, @visible,
    @gender, @birth_date, @height_cm, @goal, @program, @categories,
    @country, @city, @ready_to_relocate, @ready_to_finance,
    @primary_club_id, @metadata, @data
)
RETURNING *;

-- name: GetProfileByUserID :one
SELECT * FROM profiles WHERE user_id = @user_id;

-- name: UpdateProfile :exec
UPDATE profiles SET
    dance_styles      = @dance_styles,
    latitude          = @latitude,
    longitude         = @longitude,
    visible           = @visible,
    gender            = @gender,
    birth_date        = @birth_date,
    height_cm         = @height_cm,
    goal              = @goal,
    program           = @program,
    categories        = @categories,
    country           = @country,
    city              = @city,
    ready_to_relocate = @ready_to_relocate,
    ready_to_finance  = @ready_to_finance,
    primary_club_id   = @primary_club_id,
    metadata          = @metadata,
    data              = @data,
    updated_at        = CURRENT_TIMESTAMP
WHERE user_id = @user_id;

-- name: UpdateProfileMetadata :exec
UPDATE profiles SET
    metadata   = @metadata,
    updated_at = CURRENT_TIMESTAMP
WHERE user_id = @user_id;

-- name: GetProfilePreview :one
SELECT
    p.user_id, p.dance_styles, p.metadata, p.visible,
    p.gender, p.birth_date, p.height_cm, p.goal, p.program,
    p.categories, p.country, p.city,
    p.primary_club_id,
    u.profile_data,
    c.name AS club_name
FROM profiles p
JOIN users u ON u.id = p.user_id
LEFT JOIN clubs c ON c.id = p.primary_club_id AND c.is_active = true
WHERE p.user_id = @user_id AND p.visible = true;

-- name: SetProfilePrimaryClub :exec
UPDATE profiles SET primary_club_id = @primary_club_id, updated_at = now() WHERE user_id = @user_id;

-- name: DeleteProfile :exec
DELETE FROM profiles WHERE user_id = @user_id;

-- name: FindNearbyVisibleProfiles :many
SELECT
    p.id, p.user_id, p.dance_styles, p.metadata, p.data,
    p.latitude, p.longitude,
    p.gender, p.birth_date, p.height_cm, p.goal, p.program,
    p.categories, p.country, p.city, p.ready_to_relocate, p.ready_to_finance,
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
  AND (sqlc.narg(preferred_gender)::varchar IS NULL OR p.gender = sqlc.narg(preferred_gender))
  AND (sqlc.narg(age_min)::smallint IS NULL OR EXTRACT(YEAR FROM AGE(p.birth_date))::smallint >= sqlc.narg(age_min))
  AND (sqlc.narg(age_max)::smallint IS NULL OR EXTRACT(YEAR FROM AGE(p.birth_date))::smallint <= sqlc.narg(age_max))
  AND (sqlc.narg(height_min)::smallint IS NULL OR p.height_cm >= sqlc.narg(height_min))
  AND (sqlc.narg(height_max)::smallint IS NULL OR p.height_cm <= sqlc.narg(height_max))
  AND (sqlc.narg(preferred_goal)::varchar IS NULL OR p.goal = sqlc.narg(preferred_goal))
  AND (sqlc.narg(preferred_program)::varchar IS NULL OR p.program = sqlc.narg(preferred_program))
  AND (sqlc.narg(preferred_categories)::text[] IS NULL OR p.categories && sqlc.narg(preferred_categories))
  AND (
      sqlc.narg(preferred_city)::varchar IS NULL
      OR p.city = sqlc.narg(preferred_city)
      OR p.ready_to_relocate = true
  )
  AND (sqlc.narg(preferred_country)::varchar IS NULL OR p.country = sqlc.narg(preferred_country))
ORDER BY distance_km ASC
LIMIT @limit_val;

-- name: GetSameClubProfiles :many
SELECT
    p.id, p.user_id, p.dance_styles, p.metadata,
    p.latitude, p.longitude,
    p.gender, p.birth_date, p.height_cm, p.goal, p.program,
    p.categories, p.country, p.city, p.ready_to_relocate, p.ready_to_finance,
    cm.club_id
FROM profiles p
JOIN club_members cm ON cm.user_id = p.user_id
WHERE cm.club_id = ANY(@club_ids::uuid[])
  AND p.visible = true
  AND p.user_id != @user_id
  AND NOT (p.user_id = ANY(@exclude_ids::uuid[]))
  AND (sqlc.narg(preferred_gender)::varchar IS NULL OR p.gender = sqlc.narg(preferred_gender))
  AND (sqlc.narg(age_min)::smallint IS NULL OR EXTRACT(YEAR FROM AGE(p.birth_date))::smallint >= sqlc.narg(age_min))
  AND (sqlc.narg(age_max)::smallint IS NULL OR EXTRACT(YEAR FROM AGE(p.birth_date))::smallint <= sqlc.narg(age_max))
  AND (sqlc.narg(height_min)::smallint IS NULL OR p.height_cm >= sqlc.narg(height_min))
  AND (sqlc.narg(height_max)::smallint IS NULL OR p.height_cm <= sqlc.narg(height_max))
  AND (sqlc.narg(preferred_goal)::varchar IS NULL OR p.goal = sqlc.narg(preferred_goal))
  AND (sqlc.narg(preferred_program)::varchar IS NULL OR p.program = sqlc.narg(preferred_program))
  AND (sqlc.narg(preferred_categories)::text[] IS NULL OR p.categories && sqlc.narg(preferred_categories))
  AND (sqlc.narg(preferred_country)::varchar IS NULL OR p.country = sqlc.narg(preferred_country))
  AND (
      sqlc.narg(preferred_city)::varchar IS NULL
      OR p.city = sqlc.narg(preferred_city)
      OR p.ready_to_relocate = true
  )
  AND (sqlc.narg(wants_partner_to_finance)::varchar IS NULL OR p.ready_to_finance = sqlc.narg(wants_partner_to_finance))
ORDER BY cm.joined_at ASC
LIMIT @limit_val;

-- name: GetNearbyClubProfiles :many
SELECT
    p.id, p.user_id, p.dance_styles, p.metadata,
    p.latitude, p.longitude,
    p.gender, p.birth_date, p.height_cm, p.goal, p.program,
    p.categories, p.country, p.city, p.ready_to_relocate, p.ready_to_finance,
    cm.club_id,
    (6371 * acos(
        cos(radians(@ref_latitude::double precision)) *
        cos(radians(c.latitude)) *
        cos(radians(c.longitude) - radians(@ref_longitude::double precision)) +
        sin(radians(@ref_latitude::double precision)) *
        sin(radians(c.latitude))
    ))::double precision AS club_dist_km
FROM profiles p
JOIN club_members cm ON cm.user_id = p.user_id
JOIN clubs c ON c.id = cm.club_id
WHERE c.id != ALL(@exclude_club_ids::uuid[])
  AND p.visible = true
  AND p.user_id != @user_id
  AND NOT (p.user_id = ANY(@exclude_ids::uuid[]))
  AND (sqlc.narg(preferred_gender)::varchar IS NULL OR p.gender = sqlc.narg(preferred_gender))
  AND (sqlc.narg(age_min)::smallint IS NULL OR EXTRACT(YEAR FROM AGE(p.birth_date))::smallint >= sqlc.narg(age_min))
  AND (sqlc.narg(age_max)::smallint IS NULL OR EXTRACT(YEAR FROM AGE(p.birth_date))::smallint <= sqlc.narg(age_max))
  AND (sqlc.narg(height_min)::smallint IS NULL OR p.height_cm >= sqlc.narg(height_min))
  AND (sqlc.narg(height_max)::smallint IS NULL OR p.height_cm <= sqlc.narg(height_max))
  AND (sqlc.narg(preferred_goal)::varchar IS NULL OR p.goal = sqlc.narg(preferred_goal))
  AND (sqlc.narg(preferred_program)::varchar IS NULL OR p.program = sqlc.narg(preferred_program))
  AND (sqlc.narg(preferred_categories)::text[] IS NULL OR p.categories && sqlc.narg(preferred_categories))
  AND (sqlc.narg(preferred_country)::varchar IS NULL OR p.country = sqlc.narg(preferred_country))
  AND (
      sqlc.narg(preferred_city)::varchar IS NULL
      OR p.city = sqlc.narg(preferred_city)
      OR p.ready_to_relocate = true
  )
  AND (sqlc.narg(wants_partner_to_finance)::varchar IS NULL OR p.ready_to_finance = sqlc.narg(wants_partner_to_finance))
ORDER BY club_dist_km ASC
LIMIT @limit_val;

-- name: GetCountryWideProfiles :many
SELECT
    p.id, p.user_id, p.dance_styles, p.metadata,
    p.latitude, p.longitude,
    p.gender, p.birth_date, p.height_cm, p.goal, p.program,
    p.categories, p.country, p.city, p.ready_to_relocate, p.ready_to_finance
FROM profiles p
WHERE p.country = @country
  AND p.visible = true
  AND p.user_id != @user_id
  AND NOT (p.user_id = ANY(@exclude_ids::uuid[]))
  AND (sqlc.narg(preferred_gender)::varchar IS NULL OR p.gender = sqlc.narg(preferred_gender))
  AND (sqlc.narg(age_min)::smallint IS NULL OR EXTRACT(YEAR FROM AGE(p.birth_date))::smallint >= sqlc.narg(age_min))
  AND (sqlc.narg(age_max)::smallint IS NULL OR EXTRACT(YEAR FROM AGE(p.birth_date))::smallint <= sqlc.narg(age_max))
  AND (sqlc.narg(height_min)::smallint IS NULL OR p.height_cm >= sqlc.narg(height_min))
  AND (sqlc.narg(height_max)::smallint IS NULL OR p.height_cm <= sqlc.narg(height_max))
  AND (sqlc.narg(preferred_goal)::varchar IS NULL OR p.goal = sqlc.narg(preferred_goal))
  AND (sqlc.narg(preferred_program)::varchar IS NULL OR p.program = sqlc.narg(preferred_program))
  AND (sqlc.narg(preferred_categories)::text[] IS NULL OR p.categories && sqlc.narg(preferred_categories))
  AND (
      sqlc.narg(preferred_city)::varchar IS NULL
      OR p.city = sqlc.narg(preferred_city)
      OR p.ready_to_relocate = true
  )
  AND (sqlc.narg(wants_partner_to_finance)::varchar IS NULL OR p.ready_to_finance = sqlc.narg(wants_partner_to_finance))
ORDER BY md5(p.user_id::text || current_date::text)
LIMIT @limit_val;
