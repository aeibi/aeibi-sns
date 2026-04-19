-- name: CreatePost :one
INSERT INTO posts (
  uid,
  author_uid,
  text,
  images,
  attachments,
  tags,
  visibility,
  pinned,
  ip
)
VALUES (
  sqlc.arg(uid),
  sqlc.arg(author_uid),
  sqlc.arg(text),
  sqlc.arg(images),
  sqlc.arg(attachments),
  sqlc.arg(tags),
  sqlc.arg(visibility),
  sqlc.arg(pinned),
  sqlc.arg(ip)
)
RETURNING
  id,
  uid,
  author_uid,
  text,
  images,
  attachments,
  tags,
  comment_count,
  collection_count,
  like_count,
  pinned,
  visibility,
  latest_replied_on,
  ip,
  status,
  created_at,
  updated_at;

-- name: GetPostByUid :one
SELECT
  id,
  uid,
  author_uid,
  text,
  images,
  attachments,
  tags,
  comment_count,
  collection_count,
  like_count,
  pinned,
  visibility,
  latest_replied_on,
  ip,
  status,
  created_at,
  updated_at
FROM posts
WHERE uid = sqlc.arg(uid)
  AND status = 'NORMAL'::post_status
LIMIT 1;

-- name: GetPostsByUIDs :many
SELECT
  id,
  uid,
  author_uid,
  text,
  images,
  attachments,
  tags,
  comment_count,
  collection_count,
  like_count,
  pinned,
  visibility,
  latest_replied_on,
  ip,
  status,
  created_at,
  updated_at
FROM posts
WHERE uid = ANY(sqlc.arg(uids)::uuid[])
  AND status = 'NORMAL'::post_status;

-- name: ListPostsPublic :many
SELECT
  id,
  uid,
  author_uid,
  text,
  images,
  attachments,
  tags,
  comment_count,
  collection_count,
  like_count,
  pinned,
  visibility,
  latest_replied_on,
  ip,
  status,
  created_at,
  updated_at
FROM posts
WHERE status = 'NORMAL'::post_status
  AND visibility = 'PUBLIC'::post_visibility
  AND (created_at, uid) < (
    sqlc.arg(cursor_created_at)::timestamptz,
    sqlc.arg(cursor_id)::uuid
  )
ORDER BY created_at DESC, uid DESC
LIMIT 20;

-- name: ListPostsByAuthor :many
SELECT
  id,
  uid,
  author_uid,
  text,
  images,
  attachments,
  tags,
  comment_count,
  collection_count,
  like_count,
  pinned,
  visibility,
  latest_replied_on,
  ip,
  status,
  created_at,
  updated_at
FROM posts
WHERE status = 'NORMAL'::post_status
  AND author_uid = sqlc.arg(author_uid)
  AND (
    NOT sqlc.arg(only_public)::boolean
    OR visibility = 'PUBLIC'::post_visibility
  )
  AND (created_at, uid) < (
    sqlc.arg(cursor_created_at)::timestamptz,
    sqlc.arg(cursor_id)::uuid
  )
ORDER BY created_at DESC, uid DESC
LIMIT 20;

-- name: ListPostsByTag :many
SELECT
  id,
  uid,
  author_uid,
  text,
  images,
  attachments,
  tags,
  comment_count,
  collection_count,
  like_count,
  pinned,
  visibility,
  latest_replied_on,
  ip,
  status,
  created_at,
  updated_at
FROM posts
WHERE status = 'NORMAL'::post_status
  AND visibility = 'PUBLIC'::post_visibility
  AND tags @> ARRAY[sqlc.arg(tag_name)::text]
  AND (created_at, uid) < (
    sqlc.arg(cursor_created_at)::timestamptz,
    sqlc.arg(cursor_id)::uuid
  )
ORDER BY created_at DESC, uid DESC
LIMIT 20;

-- name: UpdatePostByUidAndAuthor :one
UPDATE posts
SET text = COALESCE(sqlc.narg(text), text),
  images = COALESCE(sqlc.narg(images)::text [], images),
  attachments = COALESCE(sqlc.narg(attachments)::text [], attachments),
  tags = COALESCE(sqlc.narg(tags)::text [], tags),
  visibility = COALESCE(
    sqlc.narg(visibility)::post_visibility,
    visibility
  ),
  pinned = COALESCE(sqlc.narg(pinned)::boolean, pinned),
  updated_at = now()
WHERE uid = sqlc.arg(uid)
  AND author_uid = sqlc.arg(author_uid)
  AND status = 'NORMAL'::post_status
RETURNING
  id,
  uid,
  author_uid,
  text,
  images,
  attachments,
  tags,
  comment_count,
  collection_count,
  like_count,
  pinned,
  visibility,
  latest_replied_on,
  ip,
  status,
  created_at,
  updated_at;

-- name: ArchivePostByUidAndAuthor :execrows
UPDATE posts
SET status = 'ARCHIVED'::post_status,
  updated_at = now()
WHERE uid = sqlc.arg(uid)
  AND author_uid = sqlc.arg(author_uid)
  AND status = 'NORMAL'::post_status;

-- name: IncrementPostCommentCount :one
UPDATE posts
SET comment_count = comment_count + 1,
    latest_replied_on = now(),
    updated_at = now()
WHERE uid = sqlc.arg(post_uid)
  AND status = 'NORMAL'::post_status
RETURNING comment_count;

-- name: DecrementPostCommentCount :one
UPDATE posts
SET comment_count = GREATEST(comment_count - 1, 0),
    updated_at = now()
WHERE uid = sqlc.arg(post_uid)
RETURNING comment_count;

-- name: GetPostCommentCount :one
SELECT comment_count
FROM posts
WHERE uid = sqlc.arg(post_uid)
  AND status = 'NORMAL'::post_status;

-- name: InsertPostLikeEdge :execrows
INSERT INTO post_likes (post_uid, user_uid)
SELECT p.uid, sqlc.arg(user_uid)
FROM posts p
WHERE p.uid = sqlc.arg(post_uid)
  AND p.status = 'NORMAL'::post_status
ON CONFLICT DO NOTHING;

-- name: DeletePostLikeEdge :execrows
DELETE FROM post_likes
WHERE post_uid = sqlc.arg(post_uid)
  AND user_uid = sqlc.arg(user_uid);

-- name: IsPostLiked :one
SELECT EXISTS (
  SELECT 1
  FROM post_likes pl
  JOIN posts p ON p.uid = pl.post_uid
  WHERE pl.post_uid = sqlc.arg(post_uid)
    AND pl.user_uid = sqlc.arg(user_uid)
    AND p.status = 'NORMAL'::post_status
) AS is_liked;

-- name: IncrementPostLikeCount :one
UPDATE posts
SET like_count = like_count + 1,
    updated_at = now()
WHERE uid = sqlc.arg(post_uid)
  AND status = 'NORMAL'::post_status
RETURNING like_count;

-- name: DecrementPostLikeCount :one
UPDATE posts
SET like_count = GREATEST(like_count - 1, 0),
    updated_at = now()
WHERE uid = sqlc.arg(post_uid)
RETURNING like_count;

-- name: GetPostLikeCount :one
SELECT like_count
FROM posts
WHERE uid = sqlc.arg(post_uid)
  AND status = 'NORMAL'::post_status;

-- name: InsertPostCollectionEdge :execrows
INSERT INTO post_collections (post_uid, user_uid)
SELECT p.uid, sqlc.arg(user_uid)
FROM posts p
WHERE p.uid = sqlc.arg(post_uid)
  AND p.status = 'NORMAL'::post_status
ON CONFLICT DO NOTHING;

-- name: DeletePostCollectionEdge :execrows
DELETE FROM post_collections
WHERE post_uid = sqlc.arg(post_uid)
  AND user_uid = sqlc.arg(user_uid);

-- name: IsPostCollected :one
SELECT EXISTS (
  SELECT 1
  FROM post_collections pc
  JOIN posts p ON p.uid = pc.post_uid
  WHERE pc.post_uid = sqlc.arg(post_uid)
    AND pc.user_uid = sqlc.arg(user_uid)
    AND p.status = 'NORMAL'::post_status
) AS is_collected;

-- name: IncrementPostCollectionCount :one
UPDATE posts
SET collection_count = collection_count + 1,
    updated_at = now()
WHERE uid = sqlc.arg(post_uid)
  AND status = 'NORMAL'::post_status
RETURNING collection_count;

-- name: DecrementPostCollectionCount :one
UPDATE posts
SET collection_count = GREATEST(collection_count - 1, 0),
    updated_at = now()
WHERE uid = sqlc.arg(post_uid)
RETURNING collection_count;

-- name: GetPostCollectionCount :one
SELECT collection_count
FROM posts
WHERE uid = sqlc.arg(post_uid)
  AND status = 'NORMAL'::post_status;

-- name: ListLikedPostUIDsByUserAndPostUIDs :many
SELECT pl.post_uid
FROM post_likes pl
JOIN posts p ON p.uid = pl.post_uid
WHERE pl.user_uid = sqlc.arg(user_uid)
  AND pl.post_uid = ANY(sqlc.arg(post_uids)::uuid[])
  AND p.status = 'NORMAL'::post_status;

-- name: ListCollectedPostUIDsByUserAndPostUIDs :many
SELECT pc.post_uid
FROM post_collections pc
JOIN posts p ON p.uid = pc.post_uid
WHERE pc.user_uid = sqlc.arg(user_uid)
  AND pc.post_uid = ANY(sqlc.arg(post_uids)::uuid[])
  AND p.status = 'NORMAL'::post_status;

-- name: ListCollectedPostRefsByUser :many
SELECT
  pc.post_uid,
  pc.created_at AS collected_at
FROM post_collections pc
JOIN posts p ON p.uid = pc.post_uid
WHERE pc.user_uid = sqlc.arg(user_uid)
  AND p.status = 'NORMAL'::post_status
  AND (pc.created_at, pc.post_uid) < (
    sqlc.arg(cursor_created_at)::timestamptz,
    sqlc.arg(cursor_id)::uuid
  )
ORDER BY pc.created_at DESC, pc.post_uid DESC
LIMIT 20;
