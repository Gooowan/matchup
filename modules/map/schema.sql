-- Map module schema

-- User locations (migrated from shared/map)
CREATE TABLE user_locations(
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id) UNIQUE,
    latitude double precision NOT NULL,
    longitude double precision NOT NULL,
    updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_user_locations_user_id ON user_locations(user_id);

CREATE INDEX idx_user_locations_coords ON user_locations(latitude, longitude);
