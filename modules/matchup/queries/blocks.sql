-- name: CreateBlock :exec
INSERT INTO blocks(blocker_id, blocked_id)
    VALUES (@blocker_id, @blocked_id)
ON CONFLICT (blocker_id, blocked_id) DO NOTHING;

-- name: DeleteBlock :exec
DELETE FROM blocks WHERE blocker_id = @blocker_id AND blocked_id = @blocked_id;

-- name: GetBlockedUserIDs :many
SELECT blocked_id FROM blocks WHERE blocker_id = @user_id;

-- name: IsBlocked :one
SELECT EXISTS(
    SELECT 1 FROM blocks
    WHERE (blocker_id = @user1_id AND blocked_id = @user2_id)
       OR (blocker_id = @user2_id AND blocked_id = @user1_id)
) AS is_blocked;
