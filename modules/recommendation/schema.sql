-- Recommendation module schema

-- Dancer profiles (1:1 with users)
CREATE TABLE profiles(
    id                 uuid          PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id            uuid          NOT NULL REFERENCES users(id) UNIQUE,
    -- Geo location
    latitude           double precision,
    longitude          double precision,
    visible            boolean       NOT NULL DEFAULT true,
    -- Dance filters (indexed, queryable)
    dance_styles       text[],
    gender             varchar(10)   NOT NULL DEFAULT '',
    birth_date         date,
    height_cm          smallint,
    goal               varchar(20)   NOT NULL DEFAULT 'hobby',
    program            varchar(20)   NOT NULL DEFAULT 'standard',
    categories         text[]        NOT NULL DEFAULT '{}',
    country            varchar(3),
    city               varchar(100),
    ready_to_relocate  boolean       DEFAULT false,
    ready_to_finance   varchar(20)   DEFAULT 'no',
    -- Non-queryable data (bio, media_urls, social links, etc.)
    metadata           jsonb         NOT NULL DEFAULT '{}',
    -- Legacy JSONB kept for rollback safety; drop after migration verified
    data               jsonb         NOT NULL DEFAULT '{}',
    created_at         timestamp     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at         timestamp     NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_profiles_user_id      ON profiles(user_id);
CREATE INDEX idx_profiles_dance_styles ON profiles USING GIN(dance_styles);
CREATE INDEX idx_profiles_coords       ON profiles(latitude, longitude);
CREATE INDEX idx_profiles_visible      ON profiles(visible) WHERE visible = true;
CREATE INDEX idx_profiles_gender       ON profiles(gender);
CREATE INDEX idx_profiles_birth_date   ON profiles(birth_date);
CREATE INDEX idx_profiles_height       ON profiles(height_cm);
CREATE INDEX idx_profiles_goal         ON profiles(goal);
CREATE INDEX idx_profiles_program      ON profiles(program);
CREATE INDEX idx_profiles_country_city ON profiles(country, city);
CREATE INDEX idx_profiles_categories   ON profiles USING GIN(categories);
CREATE INDEX idx_profiles_relocate     ON profiles(ready_to_relocate) WHERE ready_to_relocate = true;

-- User matching preferences (dedicated columns for SQL-level filtering)
CREATE TABLE user_preferences(
    id                         uuid          PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id                    uuid          NOT NULL REFERENCES users(id) UNIQUE,
    -- Filter preferences (mirror profile filter fields)
    preferred_gender           varchar(10),
    age_min                    smallint,
    age_max                    smallint,
    height_min                 smallint,
    height_max                 smallint,
    preferred_goal             varchar(20),
    preferred_program          varchar(20),
    preferred_categories       text[],
    preferred_country          varchar(3),
    preferred_city             varchar(100),
    wants_partner_to_relocate  boolean,
    wants_partner_to_finance   varchar(20),
    -- Algo metadata for Tier 2/3 (future use)
    metadata                   jsonb         NOT NULL DEFAULT '{}',
    -- Legacy JSONB kept for rollback safety
    data                       jsonb         NOT NULL DEFAULT '{}',
    created_at                 timestamp     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at                 timestamp     NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_user_preferences_user_id ON user_preferences(user_id);

-- Recommendation likes log (for Tier 2 preference model + Tier 3 collaborative filtering)
CREATE TABLE recommendation_likes_log (
    id         bigserial    PRIMARY KEY,
    user_id    uuid         NOT NULL REFERENCES users(id),
    liked_id   uuid         NOT NULL REFERENCES users(id),
    features   jsonb        NOT NULL,
    created_at timestamp    NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_rec_likes_user  ON recommendation_likes_log(user_id);
CREATE INDEX idx_rec_likes_liked ON recommendation_likes_log(liked_id);
