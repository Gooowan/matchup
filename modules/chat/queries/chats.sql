-- name: CreateChat :one
INSERT INTO chats(user1_id, user2_id)
    VALUES (@user1_id, @user2_id)
ON CONFLICT (user1_id, user2_id) WHERE club_id IS NULL DO UPDATE
    SET user1_id = EXCLUDED.user1_id
RETURNING *;

-- name: CreateClubChat :one
INSERT INTO chats(user1_id, club_id)
    VALUES (@user1_id, @club_id)
ON CONFLICT (user1_id, club_id) WHERE club_id IS NOT NULL DO UPDATE
    SET user1_id = EXCLUDED.user1_id
RETURNING *;

-- name: GetChat :one
SELECT * FROM chats WHERE id = @chat_id;

-- name: GetChatByUsers :one
SELECT * FROM chats
WHERE (user1_id = @user1_id AND user2_id = @user2_id)
   OR (user1_id = @user2_id AND user2_id = @user1_id);

-- name: ListUserChats :many
SELECT
    c.*,
    CASE WHEN c.user1_id = @user_id THEN c.user2_id ELSE c.user1_id END AS other_user_id
FROM chats c
WHERE c.user1_id = @user_id OR c.user2_id = @user_id
ORDER BY c.created_at DESC;
