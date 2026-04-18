-- name: CreateReport :one
INSERT INTO reports (
  uid,
  reporter_uid,
  report_target_type,
  target_uid,
  content
)
VALUES (
  sqlc.arg(uid),
  sqlc.arg(reporter_uid),
  sqlc.arg(report_target_type),
  sqlc.arg(target_uid),
  sqlc.arg(content)
)
RETURNING
  id,
  uid,
  reporter_uid,
  report_target_type,
  target_uid,
  content,
  status,
  created_at,
  updated_at;
