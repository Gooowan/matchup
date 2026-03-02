-- name: AdminSearchUsers :many
SELECT 
    u.id,
    u.telegram_id,
    u.email,
    u.referral_id,
    u.inviter_id,
    u.metadata,
    u.profile_data,
    u.telegram_data,
    u.created_at,
    u.role,
    u.auth_nonce,
    COUNT(*) OVER() as total_count
FROM users u
WHERE 
    (
        CASE 
            WHEN @search_term::text = '' THEN true
            ELSE (
                u.email ILIKE '%' || @search_term::text || '%' OR
                (u.profile_data->>'first_name') ILIKE '%' || @search_term::text || '%' OR
                (u.profile_data->>'last_name') ILIKE '%' || @search_term::text || '%' OR
                u.referral_id::text = @search_term::text
            )
        END
    )
ORDER BY u.created_at DESC
LIMIT @limit_val OFFSET @offset_val;

-- name: AdminGetUser :one
SELECT 
    u.id,
    u.telegram_id,
    u.email,
    u.referral_id,
    u.inviter_id,
    u.metadata,
    u.profile_data,
    u.telegram_data,
    u.created_at,
    u.role,
    u.auth_nonce,
    u.forgot_password_token,
    u.email_verification_token
FROM users u
WHERE u.id = @user_id;

-- name: AdminUpdateUser :exec
UPDATE users 
SET 
    metadata = CASE 
        WHEN @metadata::jsonb IS NOT NULL THEN COALESCE(metadata, '{}'::jsonb) || @metadata::jsonb
        ELSE metadata
    END
WHERE id = @user_id;

-- name: AdminGetTotalUsers :one
SELECT COUNT(*) as total_users
FROM users;
