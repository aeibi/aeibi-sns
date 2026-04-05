-- name: CreateReport :exec
INSERT INTO reports (
    reporter_uid,
    report_target_type,
    target_uid,
    content
  )
VALUES (@reporter_uid, @report_target_type, @target_uid, @content) ON CONFLICT (reporter_uid, report_target_type, target_uid)
WHERE status = 'NORMAL'::report_status DO NOTHING;
