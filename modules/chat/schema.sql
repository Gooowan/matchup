CREATE TABLE messages(
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    sender_id uuid NOT NULL REFERENCES users(id),
    receiver_id uuid NOT NULL REFERENCES users(id),
    type varchar(20) NOT NULL DEFAULT 'TEXT',
    content text NOT NULL,
    media_id FOREIGN KEY media(id),
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_messages_chat_created ON messages(sender_id, reciever_id, created_at);
