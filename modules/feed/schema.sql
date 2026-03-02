CREATE dance_style as ENUM(
    'Latina',
)

CREATE dance_level as ENUM(
    'Beginner',
    'Amateur',
    'Professional'
)

CREATE gender as ENUM(
    'Male',
    'Female',
    'Non-binary'
)

CREATE TABLE profiles(
    user_id uuid NOT NULL REFERENCES users(id) UNIQUE,
    styles dance_style[] not null,
    level  dance_level not null,
    height_cm int not null,
    birth_date date not null,
    gender gender not null,
    visible boolean NOT NULL DEFAULT true,
);

CREATE TABLE matches(
    from_user_id uuid NOT NULL REFERENCES users(id),
    to_user_id uuid NOT NULL REFERENCES users(id),
    match BOOLEAN NOT NULL,
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(from_user_id, to_user_id)
);

CREATE INDEX idx_matches_from_user ON matches(from_user_id, match)
WHERE match is true;

CREATE INDEX idx_matches_to_user ON matches(to_user_id, match)
WHERE match is true;