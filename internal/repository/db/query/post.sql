-- name: CreatePost :one
INSERT INTO posts (
    author,
    text,
    images,
    attachments,
    visibility,
    pinned,
    ip
  )
VALUES (
    @author,
    @text,
    COALESCE(@images::text [], '{}'::text []),
    COALESCE(@attachments::text [], '{}'::text []),
    COALESCE(
      sqlc.narg(visibility)::post_visibility,
      'PUBLIC'::post_visibility
    ),
    @pinned,
    @ip
  )
RETURNING id,
  uid;
-- name: UpsertPostTags :exec
WITH input AS (
  SELECT DISTINCT unnest(@tags::text []) AS name
),
upsert AS (
  INSERT INTO tags(name)
  SELECT name
  FROM input ON CONFLICT (name) DO
  UPDATE
  SET name = EXCLUDED.name
  RETURNING id
),
new_ids AS (
  SELECT id AS tag_id
  FROM upsert
),
del AS (
  DELETE FROM post_tags pt
  WHERE pt.post_id = @post_id
    AND NOT EXISTS (
      SELECT 1
      FROM new_ids n
      WHERE n.tag_id = pt.tag_id
    )
)
INSERT INTO post_tags (post_id, tag_id)
SELECT @post_id,
  tag_id
FROM new_ids ON CONFLICT (post_id, tag_id) DO NOTHING;
-- name: GetPostByUid :one
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
FROM posts p
  JOIN users u ON u.uid = p.author
  AND u.status = 'NORMAL'::user_status
  LEFT JOIN post_likes pl ON pl.post_uid = p.uid
  AND pl.user_uid = sqlc.narg(viewer)::uuid
  LEFT JOIN post_collections pc ON pc.post_uid = p.uid
  AND pc.user_uid = sqlc.narg(viewer)::uuid
  LEFT JOIN user_follows uf ON uf.follower_uid = sqlc.narg(viewer)::uuid
  AND uf.followee_uid = p.author
WHERE p.uid = @uid
  AND p.status = 'NORMAL'::post_status
LIMIT 1;
-- name: ListPosts :many
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
-- name: UpdatePostByUidAndAuthor :one
UPDATE posts
SET text = COALESCE(sqlc.narg(text), text),
  images = COALESCE(sqlc.narg(images)::text [], images),
  attachments = COALESCE(sqlc.narg(attachments)::text [], attachments),
  visibility = COALESCE(
    sqlc.narg(visibility)::post_visibility,
    visibility
  ),
  pinned = COALESCE(sqlc.narg(pinned)::boolean, pinned),
  updated_at = now()
WHERE uid = @uid
  AND author = @author
  AND status = 'NORMAL'::post_status
RETURNING id;
-- name: ArchivePostByUidAndAuthor :execrows
UPDATE posts
SET status = 'ARCHIVED'::post_status,
  updated_at = now()
WHERE uid = @uid
  AND author = @author
  AND status = 'NORMAL'::post_status;
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
  AND p.author = @author
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
  AND EXISTS (
    SELECT 1
    FROM post_tags pt
      JOIN tags t ON t.id = pt.tag_id
    WHERE pt.post_id = p.id
      AND t.name = @tag_name
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
