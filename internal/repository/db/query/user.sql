-- name: CreateUser :exec
INSERT INTO users (
    username,
    nickname,
    password_hash,
    email,
    avatar_url
  )
VALUES ($1, $2, $3, $4, $5);
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
WHERE uid = $1
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
WHERE username = $1
  AND status = 'NORMAL'::user_status;
-- name: UpdateUser :exec
UPDATE users
SET username = COALESCE(sqlc.narg(username), username),
  email = COALESCE(sqlc.narg(email), email),
  nickname = COALESCE(sqlc.narg(nickname), nickname),
  avatar_url = COALESCE(sqlc.narg(avatar_url), avatar_url),
  updated_at = now()
WHERE uid = $1
  AND status = 'NORMAL'::user_status;
-- name: GetUserPasswordHashByUid :one
SELECT password_hash
FROM users
WHERE uid = $1
  AND status = 'NORMAL'::user_status;
-- name: UpdateUserPasswordByUid :execrows
UPDATE users
SET password_hash = $2,
  updated_at = now()
WHERE uid = $1
  AND status = 'NORMAL'::user_status;