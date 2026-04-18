-- name: CreateFile :one
INSERT INTO files (
  url,
  name,
  content_type,
  size,
  checksum,
  uploader
)
VALUES (
  sqlc.arg(url),
  sqlc.arg(name),
  sqlc.arg(content_type),
  sqlc.arg(size),
  sqlc.arg(checksum),
  sqlc.arg(uploader)
)
RETURNING
  url,
  name,
  content_type,
  size,
  checksum,
  uploader;

-- name: GetFileByURL :one
SELECT
  url,
  name,
  content_type,
  size,
  checksum,
  uploader,
  status,
  created_at
FROM files
WHERE url = sqlc.arg(url);

-- name: GetFilesByUrls :many
SELECT
  url,
  name,
  content_type,
  size,
  checksum
FROM files
WHERE status = 'NORMAL'::file_status
  AND url = ANY(sqlc.arg(urls)::text[]);
