-- Feed module schema

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
