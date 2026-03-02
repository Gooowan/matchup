-- name: GetUsersCount :one
SELECT
    COUNT(*)
FROM
    users;

-- name: GetUser :one
SELECT
    *
FROM
    users
WHERE
    id = @user_id;

-- name: GetUserByEmail :one
SELECT
    *
FROM
    users
WHERE
    email = @email;

-- name: GetUserByReferralId :one
SELECT
    *
FROM
    users
WHERE
    referral_id = @referral_id;

-- name: GetUserByEmailVerificationToken :one
SELECT
    *
FROM
    users
WHERE
    email_verification_token = @email_verification_token;

-- name: UpdateUserEmailVerificationToken :exec
UPDATE
    users
SET
    email_verification_token = @email_verification_token
WHERE
    id = @user_id;

-- name: GetUserByForgotPasswordToken :one
SELECT
    *
FROM
    users
WHERE
    forgot_password_token = @forgot_password_token;

-- name: UpdateUserForgotPasswordToken :exec
UPDATE
    users
SET
    forgot_password_token = @forgot_password_token
WHERE
    id = @user_id;

-- name: UpdateUserProfileData :exec
UPDATE
    users
SET
    profile_data = COALESCE(profile_data, '{}'::jsonb) || @profile_data::jsonb
WHERE
    id = @user_id;

-- name: UpdateUserMetadata :exec
UPDATE
    users
SET
    metadata = COALESCE(metadata, '{}'::jsonb) || @metadata::jsonb
WHERE
    id = @user_id;

-- name: UpdateUserPassword :exec
UPDATE
    users
SET
    PASSWORD = @password
WHERE
    id = @user_id;

-- name: UpdateUserRole :exec
UPDATE
    users
SET
    ROLE = @role
WHERE
    id = @user_id;

-- name: CreateUser :one
INSERT INTO users(email, email_verification_token, PASSWORD, inviter_id, profile_data, metadata)
    VALUES (@email, @email_verification_token, @password, @inviter_id, @profile_data::jsonb, @metadata::jsonb)
RETURNING
    *;

-- name: IncrementUserNonce :one
UPDATE
    users
SET
    auth_nonce = auth_nonce + 1
WHERE
    id = @user_id
RETURNING
    auth_nonce;

