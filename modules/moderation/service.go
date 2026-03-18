package moderation

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	gen "github.com/Gooowan/matchup/modules/moderation/gen"
)

type ModerationService struct {
	Queries *gen.Queries
}

func NewModerationService(db *pgxpool.Pool) *ModerationService {
	return &ModerationService{Queries: gen.New(db)}
}

// ReportUser constructs the optional comment field and creates the report.
func (s *ModerationService) ReportUser(ctx context.Context, reporterID, reportedID pgtype.UUID, category, comment string) error {
	_, err := s.Queries.CreateReport(ctx, gen.CreateReportParams{
		ReporterID: reporterID,
		ReportedID: reportedID,
		Category:   category,
		Comment:    pgtype.Text{String: comment, Valid: comment != ""},
	})
	return err
}
