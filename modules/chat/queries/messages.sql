-- name: CreateMessage :one
INSERT INTO messages(chat_id, sender_id, type, content)
    VALUES (@chat_id, @sender_id, @type, @content)
RETURNING *;

-- name: ListMessages :many
SELECT * FROM messages
WHERE chat_id = @chat_id AND created_at < @cursor_time
ORDER BY created_at DESC
LIMIT @limit_val;

-- name: GetLatestMessage :one
SELECT * FROM messages
WHERE chat_id = @chat_id
ORDER BY created_at DESC
LIMIT 1;
