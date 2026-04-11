-- name: SearchUsers :many
SELECT u.uid,
  u.role,
  u.nickname,
  u.avatar_url,
  u.followers_count,
  u.following_count,
  u.description,
  u.status,
  u.created_at,
  COALESCE(pgroonga_score(u.tableoid, u.ctid), 0::float8)::float8 AS score
FROM users u
WHERE u.status = 'NORMAL'::user_status
  AND ARRAY [u.nickname, u.description] &@~ @query
ORDER BY score DESC,
  u.followers_count DESC,
  u.created_at DESC,
  u.uid DESC
LIMIT 20;
-- name: SuggestUsersByNicknamePrefix :many
SELECT u.uid,
  u.role,
  u.nickname,
  u.avatar_url,
  u.followers_count,
  u.following_count,
  u.description
FROM users u
WHERE u.status = 'NORMAL'::user_status
  AND u.nickname &^ @prefix
ORDER BY u.followers_count DESC,
  u.created_at DESC,
  u.uid DESC
LIMIT 10;
-- name: SearchTags :many
SELECT t.id,
  t.name,
  COALESCE(pgroonga_score(t.tableoid, t.ctid), 0::float8)::float8 AS score
FROM tags t
WHERE t.name &@~ @query
ORDER BY score DESC,
  t.name ASC
LIMIT 20;
-- name: SuggestTagsByPrefix :many
SELECT t.id,
  t.name
FROM tags t
WHERE t.name &^ @prefix
ORDER BY char_length(t.name) ASC,
  t.name ASC
LIMIT 10;
