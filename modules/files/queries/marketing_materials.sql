-- name: CreateMaterial :one
INSERT INTO media (
    name,
    file_key,
    file_size,
    content_type,
    visible
) VALUES (
    @name,
    @file_key,
    @file_size,
    @content_type,
    COALESCE(@visible, false)
) RETURNING *;

-- name: GetMaterial :one
SELECT * FROM media
WHERE id = @id AND owner_id IS NULL;

-- name: GetMaterialByKey :one
SELECT * FROM media
WHERE file_key = @file_key AND owner_id IS NULL;

-- name: ListMaterials :many
SELECT
    *,
    COUNT(*) OVER() as total_count
FROM media
WHERE owner_id IS NULL
ORDER BY created_at DESC
LIMIT @limit_val
OFFSET @offset_val;

-- name: ListVisibleMaterials :many
SELECT
    *,
    COUNT(*) OVER() as total_count
FROM media
WHERE owner_id IS NULL AND visible = true
ORDER BY created_at DESC
LIMIT @limit_val
OFFSET @offset_val;

-- name: UpdateMaterialName :exec
UPDATE media
SET
    name = @name,
    updated_at = CURRENT_TIMESTAMP
WHERE id = @id AND owner_id IS NULL;

-- name: UpdateMaterialFileKey :exec
UPDATE media
SET
    file_key = @file_key,
    updated_at = CURRENT_TIMESTAMP
WHERE id = @id AND owner_id IS NULL;

-- name: UpdateMaterialVisibility :exec
UPDATE media
SET
    visible = @visible,
    updated_at = CURRENT_TIMESTAMP
WHERE id = @id AND owner_id IS NULL;

-- name: DeleteMaterial :exec
DELETE FROM media
WHERE id = @id AND owner_id IS NULL;

