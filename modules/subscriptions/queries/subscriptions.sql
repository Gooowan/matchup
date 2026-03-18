-- name: CreateSubscription :one
INSERT INTO subscriptions (name, description, duration_days, price_cents)
VALUES (@name, @description, @duration_days, @price_cents)
RETURNING *;

-- name: GetSubscription :one
SELECT * FROM subscriptions WHERE id = @id;

-- name: ListSubscriptions :many
SELECT * FROM subscriptions WHERE is_active = true ORDER BY created_at DESC;

-- name: ListAllSubscriptions :many
SELECT * FROM subscriptions ORDER BY created_at DESC;

-- name: UpdateSubscription :exec
UPDATE subscriptions
SET name = @name,
    description = @description,
    duration_days = @duration_days,
    price_cents = @price_cents,
    is_active = @is_active,
    updated_at = CURRENT_TIMESTAMP
WHERE id = @id;

-- name: DeactivateSubscription :exec
UPDATE subscriptions SET is_active = false, updated_at = CURRENT_TIMESTAMP WHERE id = @id;
