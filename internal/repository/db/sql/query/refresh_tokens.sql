-- name: RotateRefreshToken :one
UPDATE refresh_tokens
SET token = sqlc.arg(new_token),
  expires_at = sqlc.arg(expires_at)
WHERE token = sqlc.arg(old_token)
  AND expires_at > now()
RETURNING uid;

-- name: UpsertRefreshToken :exec
INSERT INTO refresh_tokens (uid, token, expires_at)
VALUES (sqlc.arg(uid), sqlc.arg(token), sqlc.arg(expires_at)) 
ON CONFLICT (uid) DO UPDATE
SET token = sqlc.arg(token),
  expires_at = sqlc.arg(expires_at);

-- name: DeleteRefreshTokenByUid :execrows
DELETE FROM refresh_tokens
WHERE uid = sqlc.arg(uid);
