-- name: SearchPosts :many
WITH scored_posts AS (
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
    p.updated_at,
    COALESCE(pgroonga_score(p.tableoid, p.ctid), 0::float8)::float8 AS score
  FROM posts p
  WHERE p.status = 'NORMAL'::post_status
    AND p.text &@~ @query
    AND (
      p.visibility = 'PUBLIC'::post_visibility
      OR p.author = sqlc.narg(viewer)::uuid
    )
    AND (
      sqlc.narg(tag)::text IS NULL
      OR EXISTS (
        SELECT 1
        FROM post_tags pt
          JOIN tags t ON t.id = pt.tag_id
        WHERE pt.post_id = p.id
          AND t.name = sqlc.narg(tag)::text
      )
    )
),
matched_posts AS (
  SELECT *
  FROM scored_posts sp
  WHERE (
      (
        sqlc.narg(cursor_score)::float8 IS NULL
        AND sqlc.narg(cursor_created_at)::timestamptz IS NULL
        AND sqlc.narg(cursor_id)::uuid IS NULL
      )
      OR (sp.score, sp.created_at, sp.uid) < (
        sqlc.narg(cursor_score)::float8,
        sqlc.narg(cursor_created_at)::timestamptz,
        sqlc.narg(cursor_id)::uuid
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
  mp.score,
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
ORDER BY mp.score DESC,
  mp.created_at DESC,
  mp.uid DESC
LIMIT 20;
-- name: SearchUsers :many
SELECT u.uid,
  u.role,
  u.nickname,
  u.avatar_url,
  u.followers_count,
  u.following_count,
  u.description,
  u.status,
  u.created_at,
  COALESCE(pgroonga_score(u.tableoid, u.ctid), 0::float8)::float8 AS score
FROM users u
WHERE u.status = 'NORMAL'::user_status
  AND ARRAY [u.nickname, u.description] &@~ @query
ORDER BY score DESC,
  u.followers_count DESC,
  u.created_at DESC,
  u.uid DESC
LIMIT 20;
-- name: SuggestUsersByNicknamePrefix :many
SELECT u.uid,
  u.role,
  u.nickname,
  u.avatar_url,
  u.followers_count,
  u.following_count,
  u.description
FROM users u
WHERE u.status = 'NORMAL'::user_status
  AND u.nickname &^ @prefix
ORDER BY u.followers_count DESC,
  u.created_at DESC,
  u.uid DESC
LIMIT 10;
-- name: SearchTags :many
SELECT t.id,
  t.name,
  COALESCE(pgroonga_score(t.tableoid, t.ctid), 0::float8)::float8 AS score
FROM tags t
WHERE t.name &@~ @query
ORDER BY score DESC,
  t.name ASC
LIMIT 20;
-- name: SuggestTagsByPrefix :many
SELECT t.id,
  t.name
FROM tags t
WHERE t.name &^ @prefix
ORDER BY char_length(t.name) ASC,
  t.name ASC
LIMIT 10;
