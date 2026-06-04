-- Chat module schema
--
-- Architecture: polymorphic 1:1 chat table.
--   Two chat kinds share a single `chats` table distinguished by which nullable
--   columns are populated:
--
--   DM (user ↔ user):
--     user1_id + user2_id are set, club_id IS NULL.
--     Created automatically on mutual LIKE in the feed swipe flow.
--
--   Club chat (user ↔ club):
--     user1_id (the dancer who opened the thread) + club_id are set, user2_id IS NULL.
--     Created via POST /clubs/:slug/chat (idempotent create-or-get).
--     Access for the club side is granted to clubs.owner_user_id.
--
-- Known limitations (intentional deferred scope):
--   - Only the club OWNER can read/reply to club chats; staff/member roles are not
--     supported without a chat_participants table.
--   - There is no "send as club entity" — messages always carry a user sender_id.
--   - Group chat (many users in one thread) is not supported; extend to a
--     chat_participants(chat_id, participant_type, participant_id) table if needed.
--   - Unclaimed clubs can receive dancer messages, but no one can reply until
--     owner_user_id is set.
--
-- Chats. Two kinds:
--   DM:        user1_id + user2_id set, club_id NULL (created on mutual match)
--   Club chat: user1_id (the dancer) + club_id set, user2_id NULL until/unless
--              we need to pin a specific owner. The club side is answered by any
--              user who owns club_id, so unclaimed clubs can still receive chats.
CREATE TABLE chats(
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user1_id uuid NOT NULL REFERENCES users(id),
    user2_id uuid REFERENCES users(id),
    club_id uuid REFERENCES clubs(id) ON DELETE CASCADE,
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- One DM per user pair; one club chat per (dancer, club).
CREATE UNIQUE INDEX uq_chats_dm ON chats(user1_id, user2_id) WHERE club_id IS NULL;
CREATE UNIQUE INDEX uq_chats_club ON chats(user1_id, club_id) WHERE club_id IS NOT NULL;

CREATE INDEX idx_chats_user1 ON chats(user1_id);

CREATE INDEX idx_chats_user2 ON chats(user2_id);

CREATE INDEX idx_chats_club ON chats(club_id);

-- Messages
CREATE TABLE messages(
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    chat_id uuid NOT NULL REFERENCES chats(id),
    sender_id uuid NOT NULL REFERENCES users(id),
    type varchar(20) NOT NULL DEFAULT 'TEXT',
    content text NOT NULL,
    -- Moderation: NULL = visible, 'hidden' = soft-deleted by admin, 'flagged' = under review
    moderation_status varchar(20),
    deleted_at timestamp,
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_messages_chat_created ON messages(chat_id, created_at);

-- Message-level reports (reporters flag specific messages for admin review).
-- Stores a content snapshot so the original message text is preserved even if
-- the message is later deleted/hidden.
CREATE TABLE message_reports(
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    message_id uuid NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    chat_id uuid NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
    reporter_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reported_user_id uuid NOT NULL REFERENCES users(id),
    category varchar(50) NOT NULL,
    comment text,
    content_snapshot text NOT NULL,  -- message text at time of report
    status varchar(20) NOT NULL DEFAULT 'open',  -- open | resolved | dismissed
    resolved_by uuid REFERENCES users(id),
    resolved_at timestamp,
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_message_reports_status ON message_reports(status, created_at DESC);

-- Tracks the last time each participant read a chat, used for unread counts
CREATE TABLE chat_reads (
    chat_id     uuid NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
    user_id     uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    last_read_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (chat_id, user_id)
);
