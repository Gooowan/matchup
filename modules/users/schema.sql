-- Core module schema
CREATE TABLE users(
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email varchar(255) UNIQUE,

    inviter_id uuid REFERENCES users(id),
    profile_data jsonb,
    metadata jsonb,

    
    -- PRIVATE --
    role varchar(20) NOT NULL DEFAULT 'USER',
    password_hash varchar(64),

    forgot_password_token varchar(64) UNIQUE,
    email_verification_token varchar(64) UNIQUE,

    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
);


CREATE INDEX idx_users_email ON users(email);

CREATE INDEX idx_users_email_verification_token ON users(email_verification_token);

CREATE INDEX idx_users_forgot_password_token ON users(forgot_password_token);