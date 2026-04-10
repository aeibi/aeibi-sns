package service

import (
	"aeibi/api"
	"aeibi/internal/repository/db"
	"aeibi/util"
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

type ReportService struct {
	db *db.Queries
}

func NewReportService(dbx *sql.DB) *ReportService {
	return &ReportService{db: db.New(dbx)}
}

func (s *ReportService) CreateReport(ctx context.Context, uid string, req *api.CreateReportRequest) error {
	var targetType db.ReportTargetType
	switch req.ReportTargetType {
	case api.ReportTargetType_REPORT_TARGET_TYPE_POST:
		targetType = db.ReportTargetTypePOST
	case api.ReportTargetType_REPORT_TARGET_TYPE_COMMENT:
		targetType = db.ReportTargetTypeCOMMENT
	case api.ReportTargetType_REPORT_TARGET_TYPE_USER:
		targetType = db.ReportTargetTypeUSER
	default:
		return fmt.Errorf("report_target_type is invalid")
	}

	if err := s.db.CreateReport(ctx, db.CreateReportParams{
		Uid:              uuid.New(),
		ReporterUid:      util.UUID(uid),
		ReportTargetType: targetType,
		TargetUid:        util.UUID(req.TargetUid),
		Content:          req.Content,
	}); err != nil {
		return fmt.Errorf("create report: %w", err)
	}

	return nil
}
