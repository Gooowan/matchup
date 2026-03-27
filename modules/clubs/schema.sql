-- Clubs module schema

-- Dance clubs / venues (first-class entities — location anchor + SEO surface)
CREATE TABLE clubs (
    id           uuid         PRIMARY KEY DEFAULT gen_random_uuid(),
    name         varchar(255) NOT NULL,
    slug         varchar(255) UNIQUE NOT NULL,
    description  text,
    country      varchar(3)   NOT NULL,
    city         varchar(100) NOT NULL,
    address      varchar(500),
    latitude     double precision NOT NULL,
    longitude    double precision NOT NULL,
    website      varchar(500),
    phone        varchar(50),
    is_verified  boolean      DEFAULT false,
    is_active    boolean      DEFAULT true,
    metadata     jsonb        DEFAULT '{}',
    created_at   timestamp    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   timestamp    NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_clubs_coords       ON clubs(latitude, longitude);
CREATE INDEX idx_clubs_country_city ON clubs(country, city);
CREATE INDEX idx_clubs_slug         ON clubs(slug);
CREATE INDEX idx_clubs_active       ON clubs(is_active) WHERE is_active = true;
CREATE INDEX idx_clubs_verified     ON clubs(is_verified) WHERE is_verified = true;

-- Club membership (many-to-many: user <-> club)
CREATE TABLE club_members (
    club_id   uuid        NOT NULL REFERENCES clubs(id) ON DELETE CASCADE,
    user_id   uuid        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role      varchar(20) DEFAULT 'member',
    joined_at timestamp   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (club_id, user_id)
);

CREATE INDEX idx_club_members_user ON club_members(user_id);
CREATE INDEX idx_club_members_club ON club_members(club_id);
