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
-- name: ListPostsByLiker :many
SELECT p.uid,
  p.author,
  u.uid AS author_uid,
  u.nickname AS author_nickname,
  u.avatar_url AS author_avatar_url,
  p.text,
  p.images,
  p.attachments,
  p.comment_count,
  p.collection_count,
  p.like_count,
  p.pinned,
  p.visibility,
  p.latest_replied_on,
  p.ip,
  p.status,
  p.created_at,
  p.updated_at,
  true AS liked,
  (c.user_uid IS NOT NULL)::boolean AS collected,
  COALESCE(
    (
      SELECT array_agg(
          t.name
          ORDER BY t.name
        )
      FROM post_tags pt
        JOIN tags t ON t.id = pt.tag_id
      WHERE pt.post_id = p.id
    ),
    '{}'::text []
  )::text [] AS tag_names
FROM post_likes l
  JOIN posts p ON p.uid = l.post_uid
  JOIN users u ON u.uid = p.author
  AND u.status = 'NORMAL'::user_status
  LEFT JOIN post_collections c ON c.post_uid = p.uid
  AND c.user_uid = @liker
WHERE p.status = 'NORMAL'::post_status
  AND l.user_uid = @liker
  AND (
    p.visibility = 'PUBLIC'::post_visibility
    OR p.author = @liker
  )
  AND (
    (
      sqlc.narg(cursor_created_at)::timestamptz IS NULL
      AND sqlc.narg(cursor_id)::uuid IS NULL
    )
    OR (p.created_at, p.uid) < (
      sqlc.narg(cursor_created_at)::timestamptz,
      sqlc.narg(cursor_id)::uuid
    )
  )
ORDER BY p.created_at DESC,
  p.uid DESC
LIMIT 20;