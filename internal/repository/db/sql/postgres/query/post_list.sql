-- name: ListPostsPublic :many
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
  (pl.user_uid IS NOT NULL)::boolean AS liked,
  (pc.user_uid IS NOT NULL)::boolean AS collected,
  (uf.follower_uid IS NOT NULL)::boolean AS following,
  COALESCE(
    (
      SELECT array_agg(t.name ORDER BY t.name)
      FROM post_tags pt
      JOIN tags t ON t.id = pt.tag_id
      WHERE pt.post_id = p.id
    ),
    '{}'::text[]
  )::text[] AS tag_names
FROM posts p
JOIN users u ON u.uid = p.author
  AND u.status = 'NORMAL'::user_status
LEFT JOIN post_likes pl ON pl.post_uid = p.uid
  AND pl.user_uid = sqlc.narg(viewer)::uuid
LEFT JOIN post_collections pc ON pc.post_uid = p.uid
  AND pc.user_uid = sqlc.narg(viewer)::uuid
LEFT JOIN user_follows uf ON uf.follower_uid = sqlc.narg(viewer)::uuid
  AND uf.followee_uid = p.author
WHERE p.status = 'NORMAL'::post_status
  AND p.visibility = 'PUBLIC'::post_visibility
  AND (p.created_at, p.uid) < (
    sqlc.arg(cursor_created_at)::timestamptz,
    sqlc.arg(cursor_id)::uuid
  )
ORDER BY p.created_at DESC, p.uid DESC
LIMIT 20;

-- name: ListPostsByAuthor :many
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
  (pl.user_uid IS NOT NULL)::boolean AS liked,
  (pc.user_uid IS NOT NULL)::boolean AS collected,
  (uf.follower_uid IS NOT NULL)::boolean AS following,
  COALESCE(
    (
      SELECT array_agg(t.name ORDER BY t.name)
      FROM post_tags pt
      JOIN tags t ON t.id = pt.tag_id
      WHERE pt.post_id = p.id
    ),
    '{}'::text[]
  )::text[] AS tag_names
FROM posts p
JOIN users u ON u.uid = p.author
  AND u.status = 'NORMAL'::user_status
LEFT JOIN post_likes pl ON pl.post_uid = p.uid
  AND pl.user_uid = sqlc.narg(viewer)::uuid
LEFT JOIN post_collections pc ON pc.post_uid = p.uid
  AND pc.user_uid = sqlc.narg(viewer)::uuid
LEFT JOIN user_follows uf ON uf.follower_uid = sqlc.narg(viewer)::uuid
  AND uf.followee_uid = p.author
WHERE p.status = 'NORMAL'::post_status
  AND p.author = @author_uid
  AND (
    p.visibility = 'PUBLIC'::post_visibility
    OR p.author = sqlc.narg(viewer)::uuid
  )
  AND (p.created_at, p.uid) < (
    sqlc.arg(cursor_created_at)::timestamptz,
    sqlc.arg(cursor_id)::uuid
  )
ORDER BY p.created_at DESC, p.uid DESC
LIMIT 20;

-- name: ListPostsByTag :many
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
  (pl.user_uid IS NOT NULL)::boolean AS liked,
  (pc.user_uid IS NOT NULL)::boolean AS collected,
  (uf.follower_uid IS NOT NULL)::boolean AS following,
  COALESCE(
    (
      SELECT array_agg(t.name ORDER BY t.name)
      FROM post_tags pt
      JOIN tags t ON t.id = pt.tag_id
      WHERE pt.post_id = p.id
    ),
    '{}'::text[]
  )::text[] AS tag_names
FROM posts p
JOIN users u ON u.uid = p.author
  AND u.status = 'NORMAL'::user_status
LEFT JOIN post_likes pl ON pl.post_uid = p.uid
  AND pl.user_uid = sqlc.narg(viewer)::uuid
LEFT JOIN post_collections pc ON pc.post_uid = p.uid
  AND pc.user_uid = sqlc.narg(viewer)::uuid
LEFT JOIN user_follows uf ON uf.follower_uid = sqlc.narg(viewer)::uuid
  AND uf.followee_uid = p.author
WHERE p.status = 'NORMAL'::post_status
  AND (
    p.visibility = 'PUBLIC'::post_visibility
    OR p.author = sqlc.narg(viewer)::uuid
  )
  AND EXISTS (
    SELECT 1
    FROM post_tags pt
    JOIN tags t ON t.id = pt.tag_id
    WHERE pt.post_id = p.id
      AND t.name = @tag_name
  )
  AND (p.created_at, p.uid) < (
    sqlc.arg(cursor_created_at)::timestamptz,
    sqlc.arg(cursor_id)::uuid
  )
ORDER BY p.created_at DESC, p.uid DESC
LIMIT 20;
