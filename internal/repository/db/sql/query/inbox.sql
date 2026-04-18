-- name: CreateCommentInboxMessage :one
INSERT INTO inbox_comment_messages (
  uid,
  receiver_uid,
  actor_uid,
  comment_uid,
  post_uid,
  parent_comment_uid
)
VALUES (
  sqlc.arg(uid),
  sqlc.arg(receiver_uid),
  sqlc.arg(actor_uid),
  sqlc.arg(comment_uid),
  sqlc.arg(post_uid),
  sqlc.arg(parent_comment_uid)
)
RETURNING
  id,
  uid,
  receiver_uid,
  actor_uid,
  comment_uid,
  post_uid,
  parent_comment_uid,
  read_at,
  created_at;

-- name: CreateFollowInboxMessage :one
INSERT INTO inbox_follow_messages (
  uid,
  receiver_uid,
  actor_uid
)
VALUES (
  sqlc.arg(uid),
  sqlc.arg(receiver_uid),
  sqlc.arg(actor_uid)
)
RETURNING
  id,
  uid,
  receiver_uid,
  actor_uid,
  read_at,
  created_at;

-- name: FollowInboxMessageNotExists :one
SELECT NOT EXISTS (
  SELECT 1
  FROM inbox_follow_messages
  WHERE receiver_uid = sqlc.arg(receiver_uid)
    AND actor_uid = sqlc.arg(actor_uid)
    AND status = 'NORMAL'::inbox_message_status
) AS not_exists;

-- name: ListCommentInboxMessages :many
SELECT
  id,
  uid,
  receiver_uid,
  actor_uid,
  comment_uid,
  post_uid,
  parent_comment_uid,
  read_at,
  created_at
FROM inbox_comment_messages
WHERE receiver_uid = sqlc.arg(receiver_uid)
  AND status = 'NORMAL'::inbox_message_status
  AND (
    sqlc.narg(is_read)::boolean IS NULL
    OR (read_at IS NOT NULL) = sqlc.narg(is_read)::boolean
  )
  AND (created_at, uid) < (
    sqlc.arg(cursor_created_at)::timestamptz,
    sqlc.arg(cursor_id)::uuid
  )
ORDER BY created_at DESC, uid DESC
LIMIT 20;

-- name: ListFollowInboxMessages :many
SELECT
  id,
  uid,
  receiver_uid,
  actor_uid,
  read_at,
  created_at
FROM inbox_follow_messages
WHERE receiver_uid = sqlc.arg(receiver_uid)
  AND status = 'NORMAL'::inbox_message_status
  AND (
    sqlc.narg(is_read)::boolean IS NULL
    OR (read_at IS NOT NULL) = sqlc.narg(is_read)::boolean
  )
  AND (created_at, uid) < (
    sqlc.arg(cursor_created_at)::timestamptz,
    sqlc.arg(cursor_id)::uuid
  )
ORDER BY created_at DESC, uid DESC
LIMIT 20;

-- name: MarkCommentInboxMessagesReadByUIDsAndReceiver :execrows
UPDATE inbox_comment_messages
SET read_at = now()
WHERE receiver_uid = sqlc.arg(receiver_uid)
  AND uid = ANY(sqlc.arg(uids)::uuid[])
  AND status = 'NORMAL'::inbox_message_status
  AND read_at IS NULL;

-- name: MarkFollowInboxMessagesReadByUIDsAndReceiver :execrows
UPDATE inbox_follow_messages
SET read_at = now()
WHERE receiver_uid = sqlc.arg(receiver_uid)
  AND uid = ANY(sqlc.arg(uids)::uuid[])
  AND status = 'NORMAL'::inbox_message_status
  AND read_at IS NULL;

-- name: MarkAllCommentInboxMessagesReadByReceiver :execrows
UPDATE inbox_comment_messages
SET read_at = now()
WHERE receiver_uid = sqlc.arg(receiver_uid)
  AND status = 'NORMAL'::inbox_message_status
  AND read_at IS NULL;

-- name: MarkAllFollowInboxMessagesReadByReceiver :execrows
UPDATE inbox_follow_messages
SET read_at = now()
WHERE receiver_uid = sqlc.arg(receiver_uid)
  AND status = 'NORMAL'::inbox_message_status
  AND read_at IS NULL;

-- name: ArchiveCommentInboxMessageByUIDAndReceiver :execrows
UPDATE inbox_comment_messages
SET status = 'ARCHIVED'::inbox_message_status
WHERE uid = sqlc.arg(uid)
  AND receiver_uid = sqlc.arg(receiver_uid)
  AND status = 'NORMAL'::inbox_message_status;

-- name: ArchiveFollowInboxMessageByUIDAndReceiver :execrows
UPDATE inbox_follow_messages
SET status = 'ARCHIVED'::inbox_message_status
WHERE uid = sqlc.arg(uid)
  AND receiver_uid = sqlc.arg(receiver_uid)
  AND status = 'NORMAL'::inbox_message_status;

-- name: CountUnreadCommentInboxMessagesByReceiver :one
SELECT COUNT(*)::int4 AS unread_count
FROM inbox_comment_messages
WHERE receiver_uid = sqlc.arg(receiver_uid)
  AND status = 'NORMAL'::inbox_message_status
  AND read_at IS NULL;

-- name: CountUnreadFollowInboxMessagesByReceiver :one
SELECT COUNT(*)::int4 AS unread_count
FROM inbox_follow_messages
WHERE receiver_uid = sqlc.arg(receiver_uid)
  AND status = 'NORMAL'::inbox_message_status
  AND read_at IS NULL;
