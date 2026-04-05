-- name: AddFollow :one
WITH inserted AS (
  INSERT INTO user_follows (follower_uid, followee_uid)
  VALUES (@follower_uid, @followee_uid) ON CONFLICT DO NOTHING
  RETURNING 1
),
updated_following AS (
  UPDATE users
  SET following_count = following_count + 1
  WHERE uid = @follower_uid
    AND EXISTS (
      SELECT 1
      FROM inserted
    )
  RETURNING following_count
),
updated_followers AS (
  UPDATE users
  SET followers_count = followers_count + 1
  WHERE uid = @followee_uid
    AND EXISTS (
      SELECT 1
      FROM inserted
    )
  RETURNING followers_count
)
SELECT COALESCE(
    (
      SELECT following_count
      FROM updated_following
    ),
    (
      SELECT following_count
      FROM users
      WHERE users.uid = @follower_uid
    )
  )::int4 AS following_count,
  COALESCE(
    (
      SELECT followers_count
      FROM updated_followers
    ),
    (
      SELECT followers_count
      FROM users
      WHERE users.uid = @followee_uid
    )
  )::int4 AS followers_count;
-- name: RemoveFollow :one
WITH deleted AS (
  DELETE FROM user_follows
  WHERE follower_uid = @follower_uid
    AND followee_uid = @followee_uid
  RETURNING 1
),
updated_following AS (
  UPDATE users
  SET following_count = GREATEST(following_count - 1, 0)
  WHERE uid = @follower_uid
    AND EXISTS (
      SELECT 1
      FROM deleted
    )
  RETURNING following_count
),
updated_followers AS (
  UPDATE users
  SET followers_count = GREATEST(followers_count - 1, 0)
  WHERE uid = @followee_uid
    AND EXISTS (
      SELECT 1
      FROM deleted
    )
  RETURNING followers_count
)
SELECT COALESCE(
    (
      SELECT following_count
      FROM updated_following
    ),
    (
      SELECT following_count
      FROM users
      WHERE users.uid = @follower_uid
    )
  )::int4 AS following_count,
  COALESCE(
    (
      SELECT followers_count
      FROM updated_followers
    ),
    (
      SELECT followers_count
      FROM users
      WHERE users.uid = @followee_uid
    )
  )::int4 AS followers_count;
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
