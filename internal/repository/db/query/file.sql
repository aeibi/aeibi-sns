-- name: CreateFile :one
INSERT INTO files (
    url,
    name,
    content_type,
    size,
    checksum,
    uploader
  )
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING url,
  name,
  content_type,
  size,
  checksum,
  uploader;
-- name: GetFileByURL :one
SELECT url,
  name,
  content_type,
  size,
  checksum,
  uploader,
  status,
  created_at
FROM files
WHERE url = $1;
-- name: GetFilesByUrls :many
SELECT url,
  name,
  content_type,
  size,
  checksum
FROM files
WHERE status = 'NORMAL'::file_status
  AND url = ANY(@urls::text []);