-- Chat module schema

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
