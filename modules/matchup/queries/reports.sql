-- name: CreateReport :one
INSERT INTO reports(reporter_id, reported_id, category, comment)
    VALUES (@reporter_id, @reported_id, @category, @comment)
RETURNING *;

-- name: ListReportsByUser :many
SELECT * FROM reports WHERE reported_id = @user_id ORDER BY created_at DESC;
