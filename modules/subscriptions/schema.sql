CREATE TABLE subscriptions (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name          varchar(100) NOT NULL UNIQUE,
    description   text,
    duration_days int NOT NULL,
    price_cents   bigint NOT NULL DEFAULT 0,
    is_active     boolean NOT NULL DEFAULT true,
    created_at    timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE user_subscriptions (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         uuid NOT NULL REFERENCES users(id),
    subscription_id uuid NOT NULL REFERENCES subscriptions(id),
    status          varchar(20) NOT NULL DEFAULT 'active',
    started_at      timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expired_at      timestamp NOT NULL
);

CREATE INDEX idx_user_subscriptions_user ON user_subscriptions(user_id);
CREATE INDEX idx_user_subscriptions_status ON user_subscriptions(status);
CREATE INDEX idx_user_subscriptions_expired ON user_subscriptions(expired_at);
CREATE INDEX idx_user_subscriptions_status_expired ON user_subscriptions(status, expired_at);
