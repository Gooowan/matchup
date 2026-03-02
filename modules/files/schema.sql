-- Files module schema

-- Marketing materials table for admin-managed downloadable content
CREATE TABLE marketing_materials (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text NOT NULL,
    file_key text NOT NULL UNIQUE,
    file_size bigint NOT NULL,
    content_type text NOT NULL,
    visible boolean NOT NULL DEFAULT false,
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Index for efficient queries by visibility status
CREATE INDEX idx_marketing_materials_visible ON marketing_materials(visible);

-- Index for ordering by creation date
CREATE INDEX idx_marketing_materials_created_at ON marketing_materials(created_at DESC);

