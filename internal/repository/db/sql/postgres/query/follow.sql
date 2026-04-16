-- name: InsertFollowEdge :execrows
INSERT INTO user_follows (follower_uid, followee_uid)
VALUES (@follower_uid, @followee_uid)
ON CONFLICT DO NOTHING;
-- name: DeleteFollowEdge :execrows
DELETE FROM user_follows
WHERE follower_uid = @follower_uid
  AND followee_uid = @followee_uid;
-- name: IncrementFollowingCount :one
UPDATE users
SET following_count = following_count + 1
WHERE uid = @uid
RETURNING following_count;
-- name: IncrementFollowersCount :one
UPDATE users
SET followers_count = followers_count + 1
WHERE uid = @uid
RETURNING followers_count;
-- name: DecrementFollowingCount :one
UPDATE users
SET following_count = GREATEST(following_count - 1, 0)
WHERE uid = @uid
RETURNING following_count;
-- name: DecrementFollowersCount :one
UPDATE users
SET followers_count = GREATEST(followers_count - 1, 0)
WHERE uid = @uid
RETURNING followers_count;
-- name: GetFollowCounts :one
SELECT
  (SELECT u.following_count FROM users u WHERE u.uid = @follower_uid)::int4 AS following_count,
  (SELECT u.followers_count FROM users u WHERE u.uid = @followee_uid)::int4 AS followers_count;
-- name: ListFollowers :many
SELECT uf.created_at AS followed_at,
  u.uid,
  u.role,
  u.nickname,
  u.avatar_url,
  u.followers_count,
  u.following_count,
  (myf.follower_uid IS NOT NULL)::boolean AS following,
  u.status,
  u.created_at
FROM user_follows uf
  JOIN users u ON u.uid = uf.follower_uid
  AND u.status = 'NORMAL'::user_status
  LEFT JOIN user_follows myf ON myf.follower_uid = @uid
  AND myf.followee_uid = uf.follower_uid
WHERE uf.followee_uid = @uid
  AND (
    sqlc.narg(query)::text IS NULL
    OR u.nickname ILIKE '%' || sqlc.narg(query)::text || '%'
  )
  AND (
    (
      sqlc.narg(cursor_created_at)::timestamptz IS NULL
      AND sqlc.narg(cursor_id)::uuid IS NULL
    )
    OR (uf.created_at, uf.follower_uid) < (
      sqlc.narg(cursor_created_at)::timestamptz,
      sqlc.narg(cursor_id)::uuid
    )
  )
ORDER BY uf.created_at DESC,
  uf.follower_uid DESC
LIMIT 20;
-- name: ListFollowing :many
SELECT uf.created_at AS followed_at,
  u.uid,
  u.role,
  u.nickname,
  u.avatar_url,
  u.followers_count,
  u.following_count,
  u.status,
  u.created_at
FROM user_follows uf
  JOIN users u ON u.uid = uf.followee_uid
  AND u.status = 'NORMAL'::user_status
WHERE uf.follower_uid = @uid
  AND (
    sqlc.narg(query)::text IS NULL
    OR u.nickname ILIKE '%' || sqlc.narg(query)::text || '%'
  )
  AND (
    (
      sqlc.narg(cursor_created_at)::timestamptz IS NULL
      AND sqlc.narg(cursor_id)::uuid IS NULL
    )
    OR (uf.created_at, uf.followee_uid) < (
      sqlc.narg(cursor_created_at)::timestamptz,
      sqlc.narg(cursor_id)::uuid
    )
  )
ORDER BY uf.created_at DESC,
  uf.followee_uid DESC
LIMIT 20;
-- name: IsFollowing :one
SELECT EXISTS(
    SELECT 1
    FROM user_follows
    WHERE follower_uid = @follower_uid
      AND followee_uid = @followee_uid
  ) AS is_following;
