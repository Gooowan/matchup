-- name: CreateMessage :one
INSERT INTO messages(chat_id, sender_id, type, content)
    VALUES (@chat_id, @sender_id, @type, @content)
RETURNING *;

-- name: ListMessages :many
SELECT * FROM messages
WHERE chat_id = @chat_id
  AND created_at < @cursor_time
  AND deleted_at IS NULL
  AND (moderation_status IS NULL OR moderation_status != 'hidden')
ORDER BY created_at DESC
LIMIT @limit_val;

-- name: GetLatestMessage :one
SELECT * FROM messages
WHERE chat_id = @chat_id
  AND deleted_at IS NULL
  AND (moderation_status IS NULL OR moderation_status != 'hidden')
ORDER BY created_at DESC
LIMIT 1;

-- name: HideMessage :exec
UPDATE messages
SET moderation_status = 'hidden', deleted_at = NOW()
WHERE id = @id;

-- name: GetMessageByID :one
SELECT * FROM messages WHERE id = @id LIMIT 1;

-- name: CreateMessageReport :one
INSERT INTO message_reports(message_id, chat_id, reporter_id, reported_user_id, category, comment, content_snapshot)
    VALUES (@message_id, @chat_id, @reporter_id, @reported_user_id, @category, @comment, @content_snapshot)
RETURNING *;

-- name: ListMessageReports :many
SELECT mr.*, m.content AS current_content
FROM message_reports mr
JOIN messages m ON m.id = mr.message_id
WHERE mr.status = @status
ORDER BY mr.created_at DESC
LIMIT @limit_val OFFSET @offset_val;

-- name: ResolveMessageReport :exec
UPDATE message_reports
SET status = @status, resolved_by = @resolved_by, resolved_at = NOW()
WHERE id = @id;
