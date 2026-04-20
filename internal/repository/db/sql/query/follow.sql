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
  uf.created_at AS followed_at,
  u.uid,
  u.role,
  u.nickname,
  u.avatar_url,
  u.followers_count,
  u.following_count
FROM user_follows uf
JOIN users u ON u.uid = uf.follower_uid
WHERE uf.followee_uid = sqlc.arg(uid)
  AND u.status = 'NORMAL'::user_status
  AND (
    sqlc.narg(query)::text IS NULL
    OR u.nickname ILIKE '%' || sqlc.narg(query)::text || '%'
  )
  AND (uf.created_at, uf.follower_uid) < (
    sqlc.arg(cursor_created_at)::timestamptz,
    sqlc.arg(cursor_id)::uuid
  )
ORDER BY uf.created_at DESC, uf.follower_uid DESC
LIMIT 20;

-- name: ListFollowing :many
SELECT
  uf.created_at AS followed_at,
  u.uid,
  u.role,
  u.nickname,
  u.avatar_url,
  u.followers_count,
  u.following_count
FROM user_follows uf
JOIN users u ON u.uid = uf.followee_uid
WHERE uf.follower_uid = sqlc.arg(uid)
  AND u.status = 'NORMAL'::user_status
  AND (
    sqlc.narg(query)::text IS NULL
    OR u.nickname ILIKE '%' || sqlc.narg(query)::text || '%'
  )
  AND (uf.created_at, uf.followee_uid) < (
    sqlc.arg(cursor_created_at)::timestamptz,
    sqlc.arg(cursor_id)::uuid
  )
ORDER BY uf.created_at DESC, uf.followee_uid DESC
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
