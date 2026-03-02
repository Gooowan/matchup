-- Files module schema

CREATE TABLE media (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id uuid REFERENCES users(id),
    name text,
    file_key text NOT NULL UNIQUE,
    file_size bigint NOT NULL,
    content_type text NOT NULL,
    visible boolean NOT NULL DEFAULT false,
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_media_visible ON media(owner_id, visible);