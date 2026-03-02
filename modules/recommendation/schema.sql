-- Recommendation module schema

-- Dancer profiles (1:1 with users)
CREATE TABLE profiles(
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id) UNIQUE,
    dance_styles text[],
    latitude double precision,
    longitude double precision,
    visible boolean NOT NULL DEFAULT true,
    data jsonb NOT NULL DEFAULT '{}',
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_profiles_user_id ON profiles(user_id);

CREATE INDEX idx_profiles_dance_styles ON profiles USING GIN(dance_styles);

CREATE INDEX idx_profiles_coords ON profiles(latitude, longitude);

CREATE INDEX idx_profiles_visible ON profiles(visible)
WHERE
    visible = true;

-- User matching preferences
CREATE TABLE user_preferences(
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id) UNIQUE,
    data jsonb NOT NULL DEFAULT '{}',
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_user_preferences_user_id ON user_preferences(user_id);
