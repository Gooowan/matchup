-- name: CreateMatch :one
INSERT INTO matches(from_user_id, to_user_id, action)
    VALUES (@from_user_id, @to_user_id, @action)
RETURNING *;

-- name: GetMatch :one
SELECT * FROM matches
WHERE from_user_id = @from_user_id AND to_user_id = @to_user_id;

-- name: CheckMutualMatch :one
SELECT EXISTS(
    SELECT 1 FROM matches
    WHERE from_user_id = @to_user_id
      AND to_user_id = @from_user_id
      AND action = 'LIKE'
) AS is_mutual;

-- name: GetSwipedUserIDs :many
SELECT to_user_id FROM matches WHERE from_user_id = @user_id;
