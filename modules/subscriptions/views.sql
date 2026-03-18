CREATE OR REPLACE VIEW user_subscriptions_expiring_1d AS
SELECT us.id, us.user_id, us.subscription_id, us.status,
       us.started_at, us.expired_at, s.name AS subscription_name
FROM user_subscriptions us
JOIN subscriptions s ON s.id = us.subscription_id
WHERE us.status = 'active'
  AND us.expired_at > NOW()
  AND us.expired_at <= NOW() + INTERVAL '1 day';

CREATE OR REPLACE VIEW user_subscriptions_expiring_1w AS
SELECT us.id, us.user_id, us.subscription_id, us.status,
       us.started_at, us.expired_at, s.name AS subscription_name
FROM user_subscriptions us
JOIN subscriptions s ON s.id = us.subscription_id
WHERE us.status = 'active'
  AND us.expired_at > NOW()
  AND us.expired_at <= NOW() + INTERVAL '7 days';
