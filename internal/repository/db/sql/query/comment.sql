-- name: CreateComment :one
INSERT INTO post_comments (
  uid,
  post_uid,
  author_uid,
  root_uid,
  parent_uid,
  reply_to_author_uid,
  content,
  images,
  ip
)
VALUES (
  sqlc.arg(uid),
  sqlc.arg(post_uid),
  sqlc.arg(author_uid),
  sqlc.arg(root_uid),
  sqlc.arg(parent_uid),
  sqlc.arg(reply_to_author_uid),
  sqlc.arg(content),
  sqlc.arg(images),
  sqlc.arg(ip)
)
RETURNING
  id,
  uid,
  author_uid,
  post_uid,
  root_uid,
  parent_uid,
  content,
  images,
  reply_count,
  like_count,
  created_at,
  updated_at;

-- name: GetCommentByUid :one
SELECT
  id,
  uid,
  author_uid,
  post_uid,
  root_uid,
  parent_uid,
  reply_to_author_uid,
  content,
  images,
  reply_count,
  like_count,
  (parent_uid IS NULL)::boolean AS is_top_level,
  created_at,
  updated_at
FROM post_comments
WHERE uid = sqlc.arg(uid)
  AND status = 'NORMAL'::comment_status
LIMIT 1;

-- name: GetCommentsByUIDs :many
SELECT
  id,
  uid,
  author_uid,
  post_uid,
  root_uid,
  parent_uid,
  reply_to_author_uid,
  content,
  images,
  reply_count,
  like_count,
  (parent_uid IS NULL)::boolean AS is_top_level,
  created_at,
  updated_at
FROM post_comments
WHERE uid = ANY(sqlc.arg(uids)::uuid[])
  AND status = 'NORMAL'::comment_status;

-- name: ListTopComments :many
SELECT
  uid,
  author_uid,
  post_uid,
  root_uid,
  parent_uid,
  content,
  images,
  reply_count,
  like_count,
  created_at,
  updated_at
FROM post_comments
WHERE status = 'NORMAL'::comment_status
  AND post_uid = sqlc.arg(post_uid)
  AND parent_uid IS NULL
  AND (created_at, uid) < (
    sqlc.arg(cursor_created_at)::timestamptz,
    sqlc.arg(cursor_id)::uuid
  )
ORDER BY created_at DESC, uid DESC
LIMIT 20;

-- name: ListReplies :many
SELECT
  uid,
  author_uid,
  post_uid,
  root_uid,
  parent_uid,
  reply_to_author_uid,
  content,
  images,
  reply_count,
  like_count,
  created_at,
  updated_at,
  COUNT(*) OVER ()::int AS total
FROM post_comments
WHERE status = 'NORMAL'::comment_status
  AND root_uid = sqlc.arg(root_uid)
  AND root_uid <> uid
ORDER BY created_at ASC, uid ASC
LIMIT 10 OFFSET (sqlc.arg(page)::int - 1) * 10;

-- name: InsertCommentLikeEdge :execrows
INSERT INTO comment_likes (comment_uid, user_uid)
VALUES (sqlc.arg(comment_uid), sqlc.arg(user_uid))
ON CONFLICT DO NOTHING;

-- name: DeleteCommentLikeEdge :execrows
DELETE FROM comment_likes
WHERE comment_uid = sqlc.arg(comment_uid)
  AND user_uid = sqlc.arg(user_uid);

-- name: ListLikedCommentUIDsByUserAndCommentUIDs :many
SELECT comment_uid
FROM comment_likes
WHERE user_uid = sqlc.arg(user_uid)
  AND comment_uid = ANY(sqlc.arg(comment_uids)::uuid[]);

-- name: IncrementCommentLikeCount :one
UPDATE post_comments
SET like_count = like_count + 1,
    updated_at = now()
WHERE uid = sqlc.arg(comment_uid)
RETURNING like_count;

-- name: DecrementCommentLikeCount :one
UPDATE post_comments
SET like_count = GREATEST(like_count - 1, 0),
    updated_at = now()
WHERE uid = sqlc.arg(comment_uid)
RETURNING like_count;

-- name: IncrementCommentReplyCount :one
UPDATE post_comments
SET reply_count = reply_count + 1,
    updated_at = now()
WHERE uid = sqlc.arg(comment_uid)
  AND status = 'NORMAL'::comment_status
RETURNING reply_count;

-- name: DecrementCommentReplyCount :one
UPDATE post_comments
SET reply_count = GREATEST(reply_count - 1, 0),
    updated_at = now()
WHERE uid = sqlc.arg(comment_uid)
  AND status = 'NORMAL'::comment_status
RETURNING reply_count;

-- name: ArchiveCommentByUidAndAuthor :execrows
UPDATE post_comments
SET status = 'ARCHIVED'::comment_status,
  updated_at = now()
WHERE uid = sqlc.arg(uid)
  AND author_uid = sqlc.arg(author_uid)
  AND status = 'NORMAL'::comment_status;
