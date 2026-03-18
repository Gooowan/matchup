-- name: CreateUserSubscription :one
INSERT INTO user_subscriptions (user_id, subscription_id, status, started_at, expired_at)
VALUES (@user_id, @subscription_id, 'active', @started_at, @expired_at)
RETURNING *;

-- name: GetUserSubscription :one
SELECT * FROM user_subscriptions WHERE id = @id;

-- name: ListUserSubscriptions :many
SELECT us.*, s.name AS subscription_name
FROM user_subscriptions us
JOIN subscriptions s ON s.id = us.subscription_id
WHERE us.user_id = @user_id
ORDER BY us.started_at DESC;

-- name: GetActiveUserSubscription :one
SELECT us.*, s.name AS subscription_name
FROM user_subscriptions us
JOIN subscriptions s ON s.id = us.subscription_id
WHERE us.user_id = @user_id AND us.status = 'active'
ORDER BY us.expired_at DESC
LIMIT 1;

-- name: FindExpiredActiveSubscriptions :many
SELECT * FROM user_subscriptions
WHERE status = 'active' AND expired_at < NOW();

-- name: FinishExpiredSubscriptions :execrows
UPDATE user_subscriptions
SET status = 'finished'
WHERE status = 'active' AND expired_at < NOW();

-- name: ListSubscriptionsExpiring1Day :many
SELECT * FROM user_subscriptions_expiring_1d;

-- name: ListSubscriptionsExpiring1Week :many
SELECT * FROM user_subscriptions_expiring_1w;

-- name: UpdateUserSubscriptionStatus :exec
UPDATE user_subscriptions SET status = @status WHERE id = @id;
