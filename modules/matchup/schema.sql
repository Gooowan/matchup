-- MatchUp module schema

-- Dancer profiles (1:1 with users)
CREATE TABLE profiles(
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id) UNIQUE,
    dance_styles text[],
    dance_role varchar(20),
    dance_level varchar(20),
    height_cm int,
    bio text,
    birth_date date,
    gender varchar(20),
    city varchar(100),
    latitude double precision,
    longitude double precision,
    visible boolean NOT NULL DEFAULT true,
    media_urls text[],
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
    preferred_styles text[],
    preferred_role varchar(20),
    min_level varchar(20),
    max_level varchar(20),
    min_height_cm int,
    max_height_cm int,
    min_age int,
    max_age int,
    max_distance_km double precision,
    gender_preference varchar(20),
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_user_preferences_user_id ON user_preferences(user_id);

-- Swipe actions (LIKE/PASS)
CREATE TABLE matches(
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    from_user_id uuid NOT NULL REFERENCES users(id),
    to_user_id uuid NOT NULL REFERENCES users(id),
    action varchar(10) NOT NULL,
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(from_user_id, to_user_id)
);

CREATE INDEX idx_matches_from_user ON matches(from_user_id);

CREATE INDEX idx_matches_to_user ON matches(to_user_id);

CREATE INDEX idx_matches_mutual ON matches(to_user_id, from_user_id)
WHERE
    action = 'LIKE';

-- Chats (created on mutual match)
CREATE TABLE chats(
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user1_id uuid NOT NULL REFERENCES users(id),
    user2_id uuid NOT NULL REFERENCES users(id),
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user1_id, user2_id)
);

CREATE INDEX idx_chats_user1 ON chats(user1_id);

CREATE INDEX idx_chats_user2 ON chats(user2_id);

-- Messages
CREATE TABLE messages(
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    chat_id uuid NOT NULL REFERENCES chats(id),
    sender_id uuid NOT NULL REFERENCES users(id),
    type varchar(20) NOT NULL DEFAULT 'TEXT',
    content text NOT NULL,
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_messages_chat_created ON messages(chat_id, created_at);

-- Blocks
CREATE TABLE blocks(
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    blocker_id uuid NOT NULL REFERENCES users(id),
    blocked_id uuid NOT NULL REFERENCES users(id),
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(blocker_id, blocked_id)
);

CREATE INDEX idx_blocks_blocker ON blocks(blocker_id);

CREATE INDEX idx_blocks_blocked ON blocks(blocked_id);

-- Reports
CREATE TABLE reports(
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    reporter_id uuid NOT NULL REFERENCES users(id),
    reported_id uuid NOT NULL REFERENCES users(id),
    category varchar(50) NOT NULL,
    comment text,
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_reports_reporter ON reports(reporter_id);

CREATE INDEX idx_reports_reported ON reports(reported_id);

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

-- TODO: marketplace_items table (buy/sell/rent dance items)
-- TODO: marketplace_favorites table
-- TODO: marketplace_item_media table
-- TODO: dance_schools table
-- TODO: profile_school_links table (many-to-many profiles <-> schools)
