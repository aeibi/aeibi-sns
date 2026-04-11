-- name: InsertPostLikeEdge :one
WITH inserted AS (
  INSERT INTO post_likes (post_uid, user_uid)
  VALUES (@post_uid, @user_uid)
  ON CONFLICT DO NOTHING
  RETURNING 1
)
SELECT EXISTS (SELECT 1 FROM inserted);
-- name: DeletePostLikeEdge :one
WITH deleted AS (
  DELETE FROM post_likes
  WHERE post_uid = @post_uid
    AND user_uid = @user_uid
  RETURNING 1
)
SELECT EXISTS (SELECT 1 FROM deleted);
-- name: IncrementPostLikeCount :one
UPDATE posts
SET like_count = like_count + 1,
    updated_at = now()
WHERE uid = @post_uid
RETURNING like_count::int4;
-- name: DecrementPostLikeCount :one
UPDATE posts
SET like_count = GREATEST(like_count - 1, 0),
    updated_at = now()
WHERE uid = @post_uid
RETURNING like_count::int4;
-- name: GetPostLikeCount :one
SELECT like_count::int4
FROM posts
WHERE uid = @post_uid;