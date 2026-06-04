-- Core module schema
CREATE TABLE users(
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email varchar(255) UNIQUE,
    inviter_id uuid REFERENCES users(id),
    metadata jsonb,
    profile_data jsonb,
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    role varchar(20) NOT NULL DEFAULT 'USER',
    password varchar(64),
    auth_nonce int NOT NULL DEFAULT 0,
    forgot_password_token varchar(64) UNIQUE,
    forgot_password_token_expires_at timestamp,
    email_verification_token varchar(64) UNIQUE
);

CREATE INDEX idx_users_email ON users(email);

CREATE INDEX idx_users_email_verification_token ON users(email_verification_token);

CREATE INDEX idx_users_forgot_password_token ON users(forgot_password_token);

-- OAuth / external identity providers. One user can have multiple linked identities
-- (e.g. password + Google + WebAuthn later). Each provider+subject pair is unique.
CREATE TABLE user_identities(
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider varchar(32) NOT NULL,          -- e.g. 'google', 'okta', 'webauthn'
    provider_subject varchar(255) NOT NULL, -- provider's unique user identifier (sub claim)
    email varchar(255),                     -- email from the provider (informational)
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(provider, provider_subject)
);

CREATE INDEX idx_user_identities_user_id ON user_identities(user_id);

-- Seed root user (used as initial inviter / admin)
INSERT INTO users (id, email, inviter_id, metadata, profile_data, role, password, auth_nonce)
VALUES (
    '00000000-0000-0000-0000-000000000001'::uuid,
    'root@matchup.local',
    NULL,
    '{"seed": true}'::jsonb,
    '{"first_name": "Root", "last_name": "User"}'::jsonb,
    'ADMIN',
    NULL,
    0
);
