-- name: AddPostLike :one
WITH inserted AS (
  INSERT INTO post_likes (post_uid, user_uid)
  VALUES (@post_uid, @user_uid) ON CONFLICT DO NOTHING
  RETURNING 1
),
updated AS (
  UPDATE posts
  SET like_count = like_count + 1,
    updated_at = now()
  WHERE uid = @post_uid
    AND EXISTS (SELECT 1 FROM inserted)
  RETURNING like_count
)
SELECT like_count
FROM updated
UNION ALL
SELECT like_count
FROM posts
WHERE uid = @post_uid
  AND NOT EXISTS (SELECT 1 FROM updated)
LIMIT 1;
-- name: RemovePostLike :one
WITH deleted AS (
  DELETE FROM post_likes
  WHERE post_uid = @post_uid
    AND user_uid = @user_uid
  RETURNING 1
),
updated AS (
  UPDATE posts
  SET like_count = GREATEST(like_count - 1, 0),
    updated_at = now()
  WHERE uid = @post_uid
    AND EXISTS (SELECT 1 FROM deleted)
  RETURNING like_count
)
SELECT like_count
FROM updated
UNION ALL
SELECT like_count
FROM posts
WHERE uid = @post_uid
  AND NOT EXISTS (SELECT 1 FROM updated)
LIMIT 1;