-- name: CreateUser :one
INSERT INTO users (
  uid,
  username,
  nickname,
  password_hash,
  email,
  avatar_url
)
VALUES (
  sqlc.arg(uid),
  sqlc.arg(username),
  sqlc.arg(nickname),
  sqlc.arg(password_hash),
  sqlc.arg(email),
  sqlc.arg(avatar_url)
)
RETURNING
  uid,
  username,
  role,
  email,
  nickname,
  avatar_url,
  followers_count,
  following_count,
  description,
  status,
  created_at;

-- name: GetUserByUid :one
SELECT uid,
  username,
  role,
  email,
  nickname,
  avatar_url,
  followers_count,
  following_count,
  description,
  status,
  created_at
FROM users
WHERE uid = sqlc.arg(uid)
  AND status = 'NORMAL'::user_status;

-- name: GetUsersByUIDs :many
SELECT uid,
  username,
  role,
  email,
  nickname,
  avatar_url,
  followers_count,
  following_count,
  description,
  status,
  created_at
FROM users
WHERE uid = ANY(sqlc.arg(uids)::uuid[])
  AND status = 'NORMAL'::user_status;

-- name: GetUserByUsername :one
SELECT uid,
  username,
  role,
  email,
  nickname,
  avatar_url,
  followers_count,
  following_count,
  description,
  status,
  created_at,
  password_hash
FROM users
WHERE username = sqlc.arg(username)
  AND status = 'NORMAL'::user_status;

-- name: UpdateUser :one
UPDATE users
SET username = COALESCE(sqlc.narg(username), username),
  email = COALESCE(sqlc.narg(email), email),
  nickname = COALESCE(sqlc.narg(nickname), nickname),
  avatar_url = COALESCE(sqlc.narg(avatar_url), avatar_url),
  updated_at = now()
WHERE uid = sqlc.arg(uid)
  AND status = 'NORMAL'::user_status
RETURNING
  uid,
  username,
  role,
  email,
  nickname,
  avatar_url,
  followers_count,
  following_count,
  description,
  status,
  created_at;

-- name: GetUserPasswordHashByUid :one
SELECT password_hash
FROM users
WHERE uid = sqlc.arg(uid)
  AND status = 'NORMAL'::user_status;

-- name: UpdateUserPasswordByUid :execrows
UPDATE users
SET password_hash = sqlc.arg(password_hash),
  updated_at = now()
WHERE uid = sqlc.arg(uid)
  AND status = 'NORMAL'::user_status;

-- name: IncrementFollowingCount :one
UPDATE users
SET following_count = following_count + 1,
  updated_at = now()
WHERE uid = sqlc.arg(uid)
  AND status = 'NORMAL'::user_status
RETURNING following_count;

-- name: IncrementFollowersCount :one
UPDATE users
SET followers_count = followers_count + 1,
  updated_at = now()
WHERE uid = sqlc.arg(uid)
  AND status = 'NORMAL'::user_status
RETURNING followers_count;

-- name: DecrementFollowingCount :one
UPDATE users
SET following_count = GREATEST(following_count - 1, 0),
  updated_at = now()
WHERE uid = sqlc.arg(uid)
  AND status = 'NORMAL'::user_status
RETURNING following_count;

-- name: DecrementFollowersCount :one
UPDATE users
SET followers_count = GREATEST(followers_count - 1, 0),
  updated_at = now()
WHERE uid = sqlc.arg(uid)
  AND status = 'NORMAL'::user_status
RETURNING followers_count;
