CREATE TABLE blocks(
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    blocker_id uuid NOT NULL REFERENCES users(id),
    blocked_id uuid NOT NULL REFERENCES users(id),
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(blocker_id, blocked_id)
);

CREATE INDEX idx_blocks_blocker ON blocks(blocker_id);
CREATE INDEX idx_blocks_blocked ON blocks(blocked_id);

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

-- Immutable audit trail for admin actions (ban, hide message, resolve report, etc.).
-- Rows are never updated or deleted; the table is append-only.
CREATE TABLE admin_audit_log(
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    admin_id    uuid NOT NULL REFERENCES users(id),
    action      varchar(64)  NOT NULL,  -- e.g. 'ban_user', 'hide_message', 'resolve_report'
    target_type varchar(32)  NOT NULL,  -- 'user' | 'message' | 'report'
    target_id   varchar(255) NOT NULL,  -- UUID string of the affected row
    metadata    jsonb,                  -- additional context (old role, report category, etc.)
    created_at  timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_admin_audit_log_admin ON admin_audit_log(admin_id);
CREATE INDEX idx_admin_audit_log_created ON admin_audit_log(created_at DESC);