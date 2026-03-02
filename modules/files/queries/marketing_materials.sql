-- name: CreateMaterial :one
INSERT INTO marketing_materials (
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
SELECT * FROM marketing_materials
WHERE id = @id;

-- name: GetMaterialByKey :one
SELECT * FROM marketing_materials
WHERE file_key = @file_key;

-- name: ListMaterials :many
SELECT 
    *,
    COUNT(*) OVER() as total_count
FROM marketing_materials
ORDER BY created_at DESC
LIMIT @limit_val
OFFSET @offset_val;

-- name: ListVisibleMaterials :many
SELECT 
    *,
    COUNT(*) OVER() as total_count
FROM marketing_materials
WHERE visible = true
ORDER BY created_at DESC
LIMIT @limit_val
OFFSET @offset_val;

-- name: UpdateMaterialName :exec
UPDATE marketing_materials
SET 
    name = @name,
    updated_at = CURRENT_TIMESTAMP
WHERE id = @id;

-- name: UpdateMaterialVisibility :exec
UPDATE marketing_materials
SET 
    visible = @visible,
    updated_at = CURRENT_TIMESTAMP
WHERE id = @id;

-- name: DeleteMaterial :exec
DELETE FROM marketing_materials
WHERE id = @id;

