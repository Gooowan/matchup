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

func (s *ModerationService) BlockUser(ctx context.Context, blockerID, blockedID pgtype.UUID) error {
	return s.Queries.CreateBlock(ctx, gen.CreateBlockParams{
		BlockerID: blockerID,
		BlockedID: blockedID,
	})
}

func (s *ModerationService) UnblockUser(ctx context.Context, blockerID, blockedID pgtype.UUID) error {
	return s.Queries.DeleteBlock(ctx, gen.DeleteBlockParams{
		BlockerID: blockerID,
		BlockedID: blockedID,
	})
}

func (s *ModerationService) ReportUser(ctx context.Context, reporterID, reportedID pgtype.UUID, category, comment string) error {
	_, err := s.Queries.CreateReport(ctx, gen.CreateReportParams{
		ReporterID: reporterID,
		ReportedID: reportedID,
		Category:   category,
		Comment:    pgtype.Text{String: comment, Valid: comment != ""},
	})
	return err
}

func (s *ModerationService) GetBlockedIDs(ctx context.Context, userID pgtype.UUID) ([]pgtype.UUID, error) {
	return s.Queries.GetBlockedUserIDs(ctx, userID)
}

func (s *ModerationService) IsBlocked(ctx context.Context, user1, user2 pgtype.UUID) (bool, error) {
	return s.Queries.IsBlocked(ctx, gen.IsBlockedParams{
		User1ID: user1,
		User2ID: user2,
	})
}
