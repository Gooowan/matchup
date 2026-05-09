-- 1. Look up the user first
SELECT id, email, role FROM users WHERE id = '<USER_ID>'::uuid;
-- or by email:
SELECT id, email, role FROM users WHERE email = 'foo@example.com';

-- 2. Check for owned clubs (owner_user_id becomes NULL, clubs are NOT deleted)
SELECT id, name, slug FROM clubs WHERE owner_user_id = '<USER_ID>'::uuid;

-- 3. Check media to clean up from object storage manually
SELECT file_key FROM media WHERE owner_id = '<USER_ID>'::uuid;

-- 4. Delete everything in one transaction
BEGIN;

DELETE FROM user_subscriptions       WHERE user_id   = '<USER_ID>'::uuid;
DELETE FROM reports                  WHERE reporter_id = '<USER_ID>'::uuid
                                        OR reported_id = '<USER_ID>'::uuid;
DELETE FROM blocks                   WHERE blocker_id  = '<USER_ID>'::uuid
                                        OR blocked_id  = '<USER_ID>'::uuid;
DELETE FROM messages                 WHERE sender_id   = '<USER_ID>'::uuid;
DELETE FROM messages                 WHERE chat_id IN (
    SELECT id FROM chats WHERE user1_id = '<USER_ID>'::uuid OR user2_id = '<USER_ID>'::uuid
);
DELETE FROM chats                    WHERE user1_id = '<USER_ID>'::uuid
                                        OR user2_id = '<USER_ID>'::uuid;
DELETE FROM matches                  WHERE from_user_id = '<USER_ID>'::uuid
                                        OR to_user_id   = '<USER_ID>'::uuid;
DELETE FROM recommendation_likes_log WHERE user_id  = '<USER_ID>'::uuid
                                        OR liked_id = '<USER_ID>'::uuid;
DELETE FROM user_locations           WHERE user_id  = '<USER_ID>'::uuid;
DELETE FROM media                    WHERE owner_id  = '<USER_ID>'::uuid;
DELETE FROM user_preferences         WHERE user_id  = '<USER_ID>'::uuid;
DELETE FROM profiles                 WHERE user_id  = '<USER_ID>'::uuid;
DELETE FROM club_members             WHERE user_id  = '<USER_ID>'::uuid;
DELETE FROM users                    WHERE id       = '<USER_ID>'::uuid;

COMMIT;