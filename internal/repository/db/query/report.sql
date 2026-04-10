-- name: CreateReport :exec
INSERT INTO reports (
    uid,
    reporter_uid,
    report_target_type,
    target_uid,
    content
  )
VALUES (@uid, @reporter_uid, @report_target_type, @target_uid, @content) ON CONFLICT (reporter_uid, report_target_type, target_uid)
WHERE status = 'NORMAL'::report_status DO NOTHING;
