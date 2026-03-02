CREATE TABLE user_locations(
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id),
    latitude double precision NOT NULL,
    longitude double precision NOT NULL,
    active BOOLEAN DEFAULT false;
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_user_locations_user_id ON user_locations(user_id);

CREATE INDEX idx_user_locations_coords ON user_locations(latitude, longitude);
