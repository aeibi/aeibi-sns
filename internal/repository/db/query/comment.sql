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
    @uid,
    @post_uid,
    @author_uid,
    @root_uid,
    @parent_uid,
    @reply_to_author_uid,
    @content,
    COALESCE(@images::text [], '{}'::text []),
    @ip
  )
RETURNING id,
  uid;
-- name: ArchiveCommentByUidAndAuthor :execrows
UPDATE post_comments
SET status = 'ARCHIVED'::comment_status,
  updated_at = now()
WHERE uid = @uid
  AND author_uid = @author_uid
  AND status = 'NORMAL'::comment_status;
-- name: GetCommentMetaByUid :one
SELECT post_uid,
  author_uid,
  root_uid,
  content
FROM post_comments
WHERE uid = @uid
  AND status = 'NORMAL'::comment_status
LIMIT 1;
-- name: GetCommentByUid :one
SELECT c.uid,
  u.uid AS author_uid,
  u.nickname AS author_nickname,
  u.avatar_url AS author_avatar_url,
  ru.nickname AS reply_to_author_nickname,
  ru.avatar_url AS reply_to_author_avatar_url,
  c.post_uid,
  c.root_uid,
  c.parent_uid,
  c.reply_to_author_uid,
  c.content,
  c.images,
  c.reply_count,
  c.like_count,
  (cl.user_uid IS NOT NULL)::boolean AS liked,
  c.created_at,
  c.updated_at
FROM post_comments c
  JOIN users u ON u.uid = c.author_uid
  AND u.status = 'NORMAL'::user_status
  LEFT JOIN users ru ON ru.uid = c.reply_to_author_uid
  AND ru.status = 'NORMAL'::user_status
  LEFT JOIN comment_likes cl ON cl.comment_uid = c.uid
  AND cl.user_uid = sqlc.narg(viewer)::uuid
WHERE c.status = 'NORMAL'::comment_status
  AND c.uid = @uid
LIMIT 1;
-- name: InsertCommentLikeEdge :one
WITH inserted AS (
  INSERT INTO comment_likes (comment_uid, user_uid)
  VALUES (@comment_uid, @user_uid)
  ON CONFLICT DO NOTHING
  RETURNING 1
)
SELECT EXISTS (SELECT 1 FROM inserted);
-- name: DeleteCommentLikeEdge :one
WITH deleted AS (
  DELETE FROM comment_likes
  WHERE comment_uid = @comment_uid
    AND user_uid = @user_uid
  RETURNING 1
)
SELECT EXISTS (SELECT 1 FROM deleted);
-- name: IncrementCommentLikeCount :one
UPDATE post_comments
SET like_count = like_count + 1,
    updated_at = now()
WHERE uid = @comment_uid
RETURNING like_count::int4;
-- name: DecrementCommentLikeCount :one
UPDATE post_comments
SET like_count = GREATEST(like_count - 1, 0),
    updated_at = now()
WHERE uid = @comment_uid
RETURNING like_count::int4;
-- name: GetCommentLikeCount :one
SELECT like_count::int4
FROM post_comments
WHERE uid = @comment_uid;
-- name: ListTopComments :many
SELECT c.uid,
  u.uid AS author_uid,
  u.nickname AS author_nickname,
  u.avatar_url AS author_avatar_url,
  ru.nickname AS reply_to_author_nickname,
  ru.avatar_url AS reply_to_author_avatar_url,
  c.post_uid,
  c.root_uid,
  c.parent_uid,
  c.reply_to_author_uid,
  c.content,
  c.images,
  c.reply_count,
  c.like_count,
  (cl.user_uid IS NOT NULL)::boolean AS liked,
  c.created_at,
  c.updated_at
FROM post_comments c
  JOIN users u ON u.uid = c.author_uid
  AND u.status = 'NORMAL'::user_status
  LEFT JOIN users ru ON ru.uid = c.reply_to_author_uid
  AND ru.status = 'NORMAL'::user_status
  LEFT JOIN comment_likes cl ON cl.comment_uid = c.uid
  AND cl.user_uid = sqlc.narg(viewer)::uuid
WHERE c.status = 'NORMAL'::comment_status
  AND c.post_uid = @post_uid
  AND c.parent_uid IS NULL
  AND (
    (
      sqlc.narg(cursor_created_at)::timestamptz IS NULL
      AND sqlc.narg(cursor_id)::uuid IS NULL
    )
    OR (c.created_at, c.uid) < (
      sqlc.narg(cursor_created_at)::timestamptz,
      sqlc.narg(cursor_id)::uuid
    )
  )
ORDER BY c.created_at DESC,
  c.uid DESC
LIMIT 20;
-- name: ListReplies :many
SELECT c.uid,
  u.uid AS author_uid,
  u.nickname AS author_nickname,
  u.avatar_url AS author_avatar_url,
  ru.nickname AS reply_to_author_nickname,
  ru.avatar_url AS reply_to_author_avatar_url,
  c.post_uid,
  c.root_uid,
  c.parent_uid,
  c.reply_to_author_uid,
  c.content,
  c.images,
  c.reply_count,
  c.like_count,
  (cl.user_uid IS NOT NULL)::boolean AS liked,
  c.created_at,
  c.updated_at,
  COUNT(*) OVER ()::int AS total
FROM post_comments c
  JOIN users u ON u.uid = c.author_uid
  AND u.status = 'NORMAL'::user_status
  LEFT JOIN users ru ON ru.uid = c.reply_to_author_uid
  AND ru.status = 'NORMAL'::user_status
  LEFT JOIN comment_likes cl ON cl.comment_uid = c.uid
  AND cl.user_uid = sqlc.narg(viewer)::uuid
WHERE c.status = 'NORMAL'::comment_status
  AND c.root_uid = @root_uid
  AND c.root_uid <> c.uid
ORDER BY c.created_at ASC,
  c.uid ASC
LIMIT 10 OFFSET (sqlc.arg(page)::int - 1) * 10;
-- name: IncrementPostCommentCount :one
UPDATE posts
SET comment_count = comment_count + 1,
  latest_replied_on = now(),
  updated_at = now()
WHERE uid = @post_uid
  AND status = 'NORMAL'::post_status
RETURNING comment_count;
-- name: DecrementPostCommentCount :one
UPDATE posts
SET comment_count = GREATEST(comment_count - 1, 0),
  updated_at = now()
WHERE uid = @post_uid
  AND status = 'NORMAL'::post_status
RETURNING comment_count;
-- name: IncrementCommentReplyCount :one
UPDATE post_comments
SET reply_count = reply_count + 1,
  updated_at = now()
WHERE uid = @comment_uid
  AND status = 'NORMAL'::comment_status
RETURNING reply_count;
-- name: DecrementCommentReplyCount :one
UPDATE post_comments
SET reply_count = GREATEST(reply_count - 1, 0),
  updated_at = now()
WHERE uid = @comment_uid
  AND status = 'NORMAL'::comment_status
RETURNING reply_count;
