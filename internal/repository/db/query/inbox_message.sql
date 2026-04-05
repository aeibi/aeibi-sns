-- name: CreateCommentInboxMessage :one
INSERT INTO inbox_messages (
    receiver_uid,
    type,
    actor_uid,
    comment_uid,
    post_uid,
    parent_uid
  )
VALUES (
    @receiver_uid,
    'COMMENT'::message_type,
    @actor_uid,
    @comment_uid,
    @post_uid,
    @parent_uid
  )
RETURNING id,
  uid;
-- name: CreateFollowInboxMessage :execrows
INSERT INTO inbox_messages (receiver_uid, type, actor_uid)
SELECT @receiver_uid,
  'FOLLOW'::message_type,
  @actor_uid
WHERE NOT EXISTS (
    SELECT 1
    FROM inbox_messages im
    WHERE im.receiver_uid = @receiver_uid
      AND im.actor_uid = @actor_uid
      AND im.type = 'FOLLOW'::message_type
      AND im.status = 'NORMAL'::message_status
  );
-- name: ArchiveInboxMessageByUidAndReceiver :execrows
UPDATE inbox_messages
SET status = 'ARCHIVED'::message_status
WHERE uid = @uid
  AND receiver_uid = @receiver_uid
  AND status = 'NORMAL'::message_status;
-- name: ListCommentInboxMessages :many
SELECT m.uid,
  m.receiver_uid,
  m.type,
  m.is_read,
  m.actor_uid,
  u.nickname AS actor_nickname,
  u.avatar_url AS actor_avatar_url,
  m.created_at,
  m.status,
  m.comment_uid,
  c.content AS comment_content,
  m.post_uid,
  m.parent_uid,
  COALESCE(pc.content, p.text, ''::text) AS parent_content
FROM inbox_messages m
  JOIN users u ON u.uid = m.actor_uid
  AND u.status = 'NORMAL'::user_status
  LEFT JOIN post_comments c ON c.uid = m.comment_uid
  AND c.status = 'NORMAL'::comment_status
  LEFT JOIN post_comments pc ON pc.uid = m.parent_uid
  AND pc.status = 'NORMAL'::comment_status
  LEFT JOIN posts p ON p.uid = m.parent_uid
  AND p.status = 'NORMAL'::post_status
WHERE m.receiver_uid = @receiver_uid
  AND m.status = 'NORMAL'::message_status
  AND m.type = 'COMMENT'::message_type
  AND (
    sqlc.narg(is_read)::boolean IS NULL
    OR m.is_read = sqlc.narg(is_read)::boolean
  )
  AND (
    (
      sqlc.narg(cursor_created_at)::timestamptz IS NULL
      AND sqlc.narg(cursor_id)::uuid IS NULL
    )
    OR (m.created_at, m.uid) < (
      sqlc.narg(cursor_created_at)::timestamptz,
      sqlc.narg(cursor_id)::uuid
    )
  )
ORDER BY m.created_at DESC,
  m.uid DESC
LIMIT 20;
-- name: ListFollowInboxMessages :many
SELECT m.uid,
  m.receiver_uid,
  m.type,
  m.is_read,
  m.actor_uid,
  u.nickname AS actor_nickname,
  u.avatar_url AS actor_avatar_url,
  m.created_at,
  m.status
FROM inbox_messages m
  JOIN users u ON u.uid = m.actor_uid
  AND u.status = 'NORMAL'::user_status
WHERE m.receiver_uid = @receiver_uid
  AND m.status = 'NORMAL'::message_status
  AND m.type = 'FOLLOW'::message_type
  AND (
    sqlc.narg(is_read)::boolean IS NULL
    OR m.is_read = sqlc.narg(is_read)::boolean
  )
  AND (
    (
      sqlc.narg(cursor_created_at)::timestamptz IS NULL
      AND sqlc.narg(cursor_id)::uuid IS NULL
    )
    OR (m.created_at, m.uid) < (
      sqlc.narg(cursor_created_at)::timestamptz,
      sqlc.narg(cursor_id)::uuid
    )
  )
ORDER BY m.created_at DESC,
  m.uid DESC
LIMIT 20;
-- name: MarkInboxMessagesReadByUidsAndReceiver :execrows
UPDATE inbox_messages
SET is_read = true
WHERE receiver_uid = @receiver_uid
  AND uid = ANY(@uids::uuid [])
  AND status = 'NORMAL'::message_status
  AND is_read = false;
-- name: MarkAllInboxMessagesReadByReceiver :execrows
UPDATE inbox_messages
SET is_read = true
WHERE receiver_uid = @receiver_uid
  AND status = 'NORMAL'::message_status
  AND is_read = false;
-- name: CountUnreadInboxMessagesByReceiver :one
SELECT COUNT(*)::int4 AS unread_count,
  COUNT(*) FILTER (
      WHERE type = 'FOLLOW'::message_type
    )::int4 AS follow_unread_count,
  COUNT(*) FILTER (
      WHERE type = 'COMMENT'::message_type
    )::int4 AS comment_unread_count
FROM inbox_messages
WHERE receiver_uid = @receiver_uid
  AND status = 'NORMAL'::message_status
  AND is_read = false;
