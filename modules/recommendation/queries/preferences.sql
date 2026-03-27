-- name: UpsertPreferences :one
INSERT INTO user_preferences(
    user_id,
    preferred_gender, age_min, age_max, height_min, height_max,
    preferred_goal, preferred_program, preferred_categories,
    preferred_country, preferred_city,
    wants_partner_to_relocate, wants_partner_to_finance,
    metadata, data
)
VALUES (
    @user_id,
    @preferred_gender, @age_min, @age_max, @height_min, @height_max,
    @preferred_goal, @preferred_program, @preferred_categories,
    @preferred_country, @preferred_city,
    @wants_partner_to_relocate, @wants_partner_to_finance,
    @metadata, @data
)
ON CONFLICT (user_id)
    DO UPDATE SET
        preferred_gender          = EXCLUDED.preferred_gender,
        age_min                   = EXCLUDED.age_min,
        age_max                   = EXCLUDED.age_max,
        height_min                = EXCLUDED.height_min,
        height_max                = EXCLUDED.height_max,
        preferred_goal            = EXCLUDED.preferred_goal,
        preferred_program         = EXCLUDED.preferred_program,
        preferred_categories      = EXCLUDED.preferred_categories,
        preferred_country         = EXCLUDED.preferred_country,
        preferred_city            = EXCLUDED.preferred_city,
        wants_partner_to_relocate = EXCLUDED.wants_partner_to_relocate,
        wants_partner_to_finance  = EXCLUDED.wants_partner_to_finance,
        metadata                  = EXCLUDED.metadata,
        data                      = EXCLUDED.data,
        updated_at                = CURRENT_TIMESTAMP
RETURNING *;

-- name: GetPreferences :one
SELECT * FROM user_preferences WHERE user_id = @user_id;
