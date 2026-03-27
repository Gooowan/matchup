-- name: CreateClub :one
INSERT INTO clubs (name, slug, description, country, city, address, latitude, longitude, website, phone, is_verified, metadata)
VALUES (@name, @slug, @description, @country, @city, @address, @latitude, @longitude, @website, @phone, @is_verified, @metadata)
RETURNING *;

-- name: GetClubByID :one
SELECT * FROM clubs WHERE id = @id;

-- name: GetClubBySlug :one
SELECT * FROM clubs WHERE slug = @slug AND is_active = true;

-- name: ListClubs :many
SELECT * FROM clubs
WHERE is_active = true
  AND (@country::varchar IS NULL OR country = @country)
  AND (@city::varchar IS NULL OR city = @city)
ORDER BY name ASC
LIMIT @limit_val OFFSET @offset_val;

-- name: ListClubsNearby :many
SELECT *,
    (6371 * acos(
        cos(radians(@latitude::double precision)) *
        cos(radians(latitude)) *
        cos(radians(longitude) - radians(@longitude::double precision)) +
        sin(radians(@latitude::double precision)) *
        sin(radians(latitude))
    ))::double precision AS distance_km
FROM clubs
WHERE is_active = true
ORDER BY distance_km ASC
LIMIT @limit_val;

-- name: UpdateClub :exec
UPDATE clubs SET
    name        = @name,
    description = @description,
    country     = @country,
    city        = @city,
    address     = @address,
    latitude    = @latitude,
    longitude   = @longitude,
    website     = @website,
    phone       = @phone,
    metadata    = @metadata,
    updated_at  = CURRENT_TIMESTAMP
WHERE id = @id;

-- name: VerifyClub :exec
UPDATE clubs SET is_verified = true, updated_at = CURRENT_TIMESTAMP WHERE id = @id;

-- name: DeactivateClub :exec
UPDATE clubs SET is_active = false, updated_at = CURRENT_TIMESTAMP WHERE id = @id;

-- name: AdminListClubs :many
SELECT * FROM clubs
ORDER BY created_at DESC
LIMIT @limit_val OFFSET @offset_val;

-- name: JoinClub :exec
INSERT INTO club_members (club_id, user_id, role)
VALUES (@club_id, @user_id, 'member')
ON CONFLICT (club_id, user_id) DO NOTHING;

-- name: LeaveClub :exec
DELETE FROM club_members WHERE club_id = @club_id AND user_id = @user_id;

-- name: GetUserClubs :many
SELECT c.* FROM clubs c
JOIN club_members cm ON cm.club_id = c.id
WHERE cm.user_id = @user_id AND c.is_active = true
ORDER BY cm.joined_at ASC;

-- name: ListClubMembers :many
SELECT
    cm.user_id,
    cm.role,
    cm.joined_at,
    p.gender,
    p.goal,
    p.program,
    p.categories,
    p.country,
    p.city,
    p.metadata
FROM club_members cm
JOIN profiles p ON p.user_id = cm.user_id
WHERE cm.club_id = @club_id AND p.visible = true
ORDER BY cm.joined_at DESC
LIMIT @limit_val OFFSET @offset_val;

-- name: GetClubMemberCount :one
SELECT COUNT(*)::int FROM club_members WHERE club_id = @club_id;

-- name: IsClubMember :one
SELECT EXISTS(
    SELECT 1 FROM club_members WHERE club_id = @club_id AND user_id = @user_id
) AS is_member;
