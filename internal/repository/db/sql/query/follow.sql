-- name: InsertFollowEdge :execrows
INSERT INTO user_follows (follower_uid, followee_uid)
VALUES (sqlc.arg(follower_uid), sqlc.arg(followee_uid))
ON CONFLICT DO NOTHING;

-- name: DeleteFollowEdge :execrows
DELETE FROM user_follows
WHERE follower_uid = sqlc.arg(follower_uid)
  AND followee_uid = sqlc.arg(followee_uid);

-- name: ListFollowers :many
SELECT
  created_at AS followed_at,
  follower_uid AS uid
FROM user_follows
WHERE followee_uid = sqlc.arg(uid)
  AND (created_at, follower_uid) < (
    sqlc.arg(cursor_created_at)::timestamptz,
    sqlc.arg(cursor_id)::uuid
  )
ORDER BY created_at DESC, follower_uid DESC
LIMIT 20;

-- name: ListFollowing :many
SELECT
  created_at AS followed_at,
  followee_uid AS uid
FROM user_follows
WHERE follower_uid = sqlc.arg(uid)
  AND (created_at, followee_uid) < (
    sqlc.arg(cursor_created_at)::timestamptz,
    sqlc.arg(cursor_id)::uuid
  )
ORDER BY created_at DESC, followee_uid DESC
LIMIT 20;

-- name: IsFollowing :one
SELECT EXISTS (
  SELECT 1
  FROM user_follows
  WHERE follower_uid = sqlc.arg(follower_uid)
    AND followee_uid = sqlc.arg(followee_uid)
) AS is_following;

-- name: ListFollowingUIDsByFollowerAndFolloweeUIDs :many
SELECT followee_uid
FROM user_follows
WHERE follower_uid = sqlc.arg(follower_uid)
  AND followee_uid = ANY(sqlc.arg(followee_uids)::uuid[]);
