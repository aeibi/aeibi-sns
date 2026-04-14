-- name: CreatePost :one
INSERT INTO posts (
    uid,
    author,
    text,
    images,
    attachments,
    visibility,
    pinned,
    ip
  )
VALUES (
    @uid,
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
-- name: InsertTagsIfNotExists :exec
WITH input AS (
  SELECT DISTINCT unnest(@tags::text[]) AS name
)
INSERT INTO tags (name)
SELECT name
FROM input
ON CONFLICT (name) DO NOTHING;
-- name: DeletePostTagsNotInNames :exec
DELETE FROM post_tags pt
WHERE pt.post_id = @post_id
  AND NOT EXISTS (
    SELECT 1
    FROM tags t
    JOIN (
      SELECT DISTINCT unnest(@tags::text[]) AS name
    ) i ON i.name = t.name
    WHERE t.id = pt.tag_id
  );
-- name: InsertPostTagsByNames :exec
WITH input AS (
  SELECT DISTINCT unnest(@tags::text[]) AS name
)
INSERT INTO post_tags (post_id, tag_id)
SELECT @post_id, t.id
FROM tags t
JOIN input i ON i.name = t.name
ON CONFLICT (post_id, tag_id) DO NOTHING;
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
-- name: GetPostSearchExtrasByUids :many
WITH input AS (
  SELECT DISTINCT ON (x.uid) x.uid,
    x.ord
  FROM unnest(@uids::uuid []) WITH ORDINALITY AS x(uid, ord)
  ORDER BY x.uid,
    x.ord
)
SELECT p.uid,
  u.nickname AS author_nickname,
  u.avatar_url AS author_avatar_url,
  (uf.follower_uid IS NOT NULL)::boolean AS is_following,
  (pl.user_uid IS NOT NULL)::boolean AS liked,
  (pc.user_uid IS NOT NULL)::boolean AS collected
FROM input i
  JOIN posts p ON p.uid = i.uid
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
ORDER BY i.ord;
-- name: ListPosts :many
WITH filtered_posts AS (
  SELECT p.id,
    p.uid,
    p.author,
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
    p.updated_at
  FROM posts p
  WHERE p.status = 'NORMAL'::post_status
    AND (
      p.visibility = 'PUBLIC'::post_visibility
      OR p.author = sqlc.narg(viewer)::uuid
    )
    AND (
      sqlc.narg(author_uid)::uuid IS NULL
      OR p.author = sqlc.narg(author_uid)::uuid
    )
    AND (
      sqlc.narg(tag_name)::text IS NULL
      OR EXISTS (
        SELECT 1
        FROM post_tags pt
          JOIN tags t ON t.id = pt.tag_id
        WHERE pt.post_id = p.id
          AND t.name = sqlc.narg(tag_name)::text
      )
    )
),
matched_posts AS (
  SELECT *
  FROM filtered_posts fp
  WHERE (
      (
        sqlc.narg(cursor_created_at)::timestamptz IS NULL
        AND sqlc.narg(cursor_id)::uuid IS NULL
      )
      OR (
        (fp.created_at, fp.uid) < (
          sqlc.narg(cursor_created_at)::timestamptz,
          sqlc.narg(cursor_id)::uuid
        )
      )
    )
)
SELECT mp.uid,
  mp.author,
  u.uid AS author_uid,
  u.nickname AS author_nickname,
  u.avatar_url AS author_avatar_url,
  mp.text,
  mp.images,
  mp.attachments,
  mp.comment_count,
  mp.collection_count,
  mp.like_count,
  mp.pinned,
  mp.visibility,
  mp.latest_replied_on,
  mp.ip,
  mp.status,
  mp.created_at,
  mp.updated_at,
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
      WHERE pt.post_id = mp.id
    ),
    '{}'::text []
  )::text [] AS tag_names
FROM matched_posts mp
  JOIN users u ON u.uid = mp.author
  AND u.status = 'NORMAL'::user_status
  LEFT JOIN post_likes pl ON pl.post_uid = mp.uid
  AND pl.user_uid = sqlc.narg(viewer)::uuid
  LEFT JOIN post_collections pc ON pc.post_uid = mp.uid
  AND pc.user_uid = sqlc.narg(viewer)::uuid
  LEFT JOIN user_follows uf ON uf.follower_uid = sqlc.narg(viewer)::uuid
  AND uf.followee_uid = mp.author
ORDER BY mp.created_at DESC,
  mp.uid DESC
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
