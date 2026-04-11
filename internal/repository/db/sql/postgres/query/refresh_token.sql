-- name: GetRefreshToken :one
SELECT uid,
  token
FROM refresh_tokens
WHERE token = $1
  AND expires_at > now();
-- name: UpsertRefreshToken :exec
INSERT INTO refresh_tokens (uid, token, expires_at)
VALUES ($1, $2, $3) ON CONFLICT (uid) DO
UPDATE
SET token = EXCLUDED.token,
  expires_at = EXCLUDED.expires_at;
-- name: DeleteRefreshTokenByUid :execrows
DELETE FROM refresh_tokens
WHERE uid = $1;