-- name: FindIdentity :one
SELECT ui.user_id
FROM user_identities ui
WHERE ui.provider = @provider
  AND ui.provider_subject = @provider_subject
LIMIT 1;

-- name: CreateIdentity :exec
INSERT INTO user_identities(user_id, provider, provider_subject, email)
VALUES (@user_id, @provider, @provider_subject, @email);
