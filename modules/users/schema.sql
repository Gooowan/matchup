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
    email_verification_token varchar(64) UNIQUE
);

CREATE INDEX idx_users_email ON users(email);

CREATE INDEX idx_users_email_verification_token ON users(email_verification_token);

CREATE INDEX idx_users_forgot_password_token ON users(forgot_password_token);
