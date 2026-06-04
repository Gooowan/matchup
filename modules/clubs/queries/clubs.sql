-- name: CreateClub :one
INSERT INTO clubs (name, slug, description, country, city, address, latitude, longitude, website, phone, is_verified, metadata, working_hours)
VALUES (@name, @slug, @description, @country, @city, @address, @latitude, @longitude, @website, @phone, @is_verified, @metadata, @working_hours)
RETURNING *;

-- name: GetClubByID :one
SELECT * FROM clubs WHERE id = @id;

-- name: GetClubBySlug :one
SELECT * FROM clubs WHERE slug = @slug AND is_active = true;

-- name: ListClubs :many
SELECT * FROM clubs
WHERE is_active = true
  AND (NULLIF(@country::varchar, '') IS NULL OR country = @country)
  AND (NULLIF(@city::varchar, '') IS NULL OR city ILIKE '%' || @city || '%')
  AND (NULLIF(@q::varchar, '') IS NULL OR name ILIKE '%' || @q || '%' OR city ILIKE '%' || @q || '%' OR address ILIKE '%' || @q || '%')
ORDER BY is_verified DESC, name ASC
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
    p.account_type,
    p.gender,
    p.birth_date,
    p.goal,
    p.program,
    p.categories,
    p.country,
    p.city,
    p.metadata,
    u.profile_data
FROM club_members cm
JOIN profiles p ON p.user_id = cm.user_id
JOIN users u ON u.id = cm.user_id
WHERE cm.club_id = @club_id AND p.visible = true
ORDER BY cm.joined_at DESC
LIMIT @limit_val OFFSET @offset_val;

-- name: GetClubMemberCount :one
SELECT COUNT(*)::int FROM club_members WHERE club_id = @club_id;

-- name: CountUserClubMemberships :one
SELECT COUNT(*)::int FROM club_members WHERE user_id = @user_id;

-- name: CountTrainerClubs :one
SELECT COUNT(*)::int FROM club_trainers WHERE trainer_user_id = @trainer_user_id;

-- name: IsClubMember :one
SELECT EXISTS(
    SELECT 1 FROM club_members WHERE club_id = @club_id AND user_id = @user_id
) AS is_member;

-- name: GetClubOwner :one
SELECT owner_user_id FROM clubs WHERE id = @id;

-- name: ClaimClub :one
UPDATE clubs
SET owner_user_id = @owner_user_id, updated_at = CURRENT_TIMESTAMP
WHERE id = @id AND owner_user_id IS NULL
RETURNING *;

-- name: ManageClub :exec
UPDATE clubs
SET name          = COALESCE(NULLIF(@name, ''), name),
    description   = @description,
    address       = @address,
    phone         = @phone,
    website       = @website,
    working_hours = @working_hours,
    metadata      = CASE
                      WHEN @logo_url::text != '' THEN
                        COALESCE(metadata, '{}')::jsonb || jsonb_build_object('logo_url', @logo_url::text)
                      ELSE metadata
                    END,
    latitude      = CASE WHEN @latitude::double precision != 0 THEN @latitude ELSE latitude END,
    longitude     = CASE WHEN @longitude::double precision != 0 THEN @longitude ELSE longitude END,
    updated_at    = CURRENT_TIMESTAMP
WHERE id = @id AND owner_user_id = @owner_user_id;

-- name: ListOwnedClubs :many
SELECT * FROM clubs WHERE owner_user_id = @owner_user_id AND is_active = true ORDER BY name ASC;

-- name: ListClubDancers :many
-- Dancers (account_type dancer/parent) who are members of the club.
SELECT
    cm.user_id,
    cm.role,
    cm.joined_at,
    p.account_type,
    p.gender,
    p.birth_date,
    p.goal,
    p.program,
    p.categories,
    p.country,
    p.city,
    p.metadata,
    u.profile_data
FROM club_members cm
JOIN profiles p ON p.user_id = cm.user_id
JOIN users u ON u.id = cm.user_id
WHERE cm.club_id = @club_id
  AND p.visible = true
  AND p.account_type IN ('dancer', 'parent')
ORDER BY cm.joined_at DESC
LIMIT @limit_val OFFSET @offset_val;

-- name: AddClubTrainer :exec
INSERT INTO club_trainers (club_id, trainer_user_id)
VALUES (@club_id, @trainer_user_id)
ON CONFLICT (club_id, trainer_user_id) DO NOTHING;

-- name: RemoveClubTrainer :exec
DELETE FROM club_trainers WHERE club_id = @club_id AND trainer_user_id = @trainer_user_id;

-- name: ListClubTrainers :many
SELECT
    ct.trainer_user_id,
    ct.joined_at,
    p.gender,
    p.categories,
    p.metadata,
    u.profile_data
FROM club_trainers ct
JOIN profiles p ON p.user_id = ct.trainer_user_id
JOIN users u ON u.id = ct.trainer_user_id
WHERE ct.club_id = @club_id AND p.visible = true
ORDER BY ct.joined_at ASC
LIMIT @limit_val OFFSET @offset_val;

-- name: ListTrainerClubs :many
SELECT c.* FROM clubs c
JOIN club_trainers ct ON ct.club_id = c.id
WHERE ct.trainer_user_id = @trainer_user_id AND c.is_active = true
ORDER BY ct.joined_at ASC;

-- name: EnrollTrainerStudent :exec
INSERT INTO trainer_students (trainer_user_id, dancer_user_id)
VALUES (@trainer_user_id, @dancer_user_id)
ON CONFLICT (trainer_user_id, dancer_user_id) DO NOTHING;

-- name: UnenrollTrainerStudent :exec
DELETE FROM trainer_students
WHERE trainer_user_id = @trainer_user_id AND dancer_user_id = @dancer_user_id;

-- name: ListTrainerStudents :many
SELECT
    ts.dancer_user_id,
    ts.enrolled_at,
    p.gender,
    p.birth_date,
    p.goal,
    p.program,
    p.categories,
    p.country,
    p.city,
    p.metadata,
    u.profile_data
FROM trainer_students ts
JOIN profiles p ON p.user_id = ts.dancer_user_id
JOIN users u ON u.id = ts.dancer_user_id
WHERE ts.trainer_user_id = @trainer_user_id AND p.visible = true
ORDER BY ts.enrolled_at DESC;

-- name: ListDancerTrainers :many
SELECT
    ts.trainer_user_id,
    ts.enrolled_at,
    p.gender,
    p.categories,
    p.metadata,
    u.profile_data
FROM trainer_students ts
JOIN profiles p ON p.user_id = ts.trainer_user_id
JOIN users u ON u.id = ts.trainer_user_id
WHERE ts.dancer_user_id = @dancer_user_id AND p.visible = true
ORDER BY ts.enrolled_at ASC;
