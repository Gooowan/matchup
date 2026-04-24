-- name: InsertLikeLog :exec
INSERT INTO recommendation_likes_log (user_id, liked_id, features)
VALUES (@user_id, @liked_id, @features);

-- name: GetLikeHistory :many
SELECT features FROM recommendation_likes_log
WHERE user_id = @user_id
ORDER BY created_at DESC
LIMIT 50;

-- name: GetSimilarUsers :many
SELECT other.user_id, COUNT(*)::bigint AS overlap
FROM recommendation_likes_log AS other
WHERE other.liked_id IN (
    SELECT self.liked_id FROM recommendation_likes_log self WHERE self.user_id = @user_id
)
  AND other.user_id != @user_id
GROUP BY other.user_id
HAVING COUNT(*) >= 2
ORDER BY overlap DESC
LIMIT 20;

-- name: GetProfilesLikedBySimilarUsers :many
SELECT DISTINCT rll.liked_id
FROM recommendation_likes_log rll
WHERE rll.user_id = ANY(@similar_user_ids::uuid[])
  AND rll.liked_id NOT IN (
      SELECT self.liked_id FROM recommendation_likes_log self WHERE self.user_id = @user_id
  )
  AND rll.liked_id != @user_id;

-- name: GetProfilesByUserIDs :many
SELECT
    p.id, p.user_id, p.latitude, p.longitude, p.dance_styles, p.metadata,
    p.gender, p.birth_date, p.height_cm, p.goal, p.program,
    p.categories, p.country, p.city, p.ready_to_relocate, p.ready_to_finance
FROM profiles p
WHERE p.user_id = ANY(@user_ids::uuid[])
  AND p.visible = true;
