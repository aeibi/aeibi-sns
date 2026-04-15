-- name: InsertPostCollectionEdge :execrows
INSERT INTO post_collections (post_uid, user_uid)
VALUES (@post_uid, @user_uid)
ON CONFLICT DO NOTHING;
-- name: DeletePostCollectionEdge :execrows
DELETE FROM post_collections
WHERE post_uid = @post_uid
  AND user_uid = @user_uid;
-- name: IncrementPostCollectionCount :one
UPDATE posts
SET collection_count = collection_count + 1,
    updated_at = now()
WHERE uid = @post_uid
RETURNING collection_count::int4;
-- name: DecrementPostCollectionCount :one
UPDATE posts
SET collection_count = GREATEST(collection_count - 1, 0),
    updated_at = now()
WHERE uid = @post_uid
RETURNING collection_count::int4;
-- name: GetPostCollectionCount :one
SELECT collection_count::int4
FROM posts
WHERE uid = @post_uid;
-- name: ListPostsByCollector :many
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
  true AS collected,
  (pl.user_uid IS NOT NULL)::boolean AS liked,
  (uf.follower_uid IS NOT NULL)::boolean AS following,
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
FROM post_collections c
  JOIN posts p ON p.uid = c.post_uid
  JOIN users u ON u.uid = p.author
  AND u.status = 'NORMAL'::user_status
  LEFT JOIN post_likes pl ON pl.post_uid = p.uid
  AND pl.user_uid = @collector
  LEFT JOIN user_follows uf ON uf.follower_uid = @collector
  AND uf.followee_uid = p.author
WHERE p.status = 'NORMAL'::post_status
  AND c.user_uid = @collector
  AND (
    p.visibility = 'PUBLIC'::post_visibility
    OR p.author = @collector
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
