package moderation

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	gen "github.com/Gooowan/matchup/modules/moderation/gen"
)

type ModerationService struct {
	DB      *pgxpool.Pool
	Queries *gen.Queries
}

func NewModerationService(db *pgxpool.Pool) *ModerationService {
	return &ModerationService{DB: db, Queries: gen.New(db)}
}

type ReportRow struct {
	ID            string    `json:"id"`
	ReporterID    string    `json:"reporter_id"`
	ReporterEmail string    `json:"reporter_email"`
	ReportedID    string    `json:"reported_id"`
	ReportedEmail string    `json:"reported_email"`
	Category      string    `json:"category"`
	Comment       string    `json:"comment"`
	CreatedAt     time.Time `json:"created_at"`
}

func (s *ModerationService) ListAllReports(ctx context.Context, limit int) ([]ReportRow, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT r.id, r.reporter_id, u1.email, r.reported_id, u2.email, r.category,
		       COALESCE(r.comment, ''), r.created_at
		FROM reports r
		LEFT JOIN users u1 ON u1.id = r.reporter_id
		LEFT JOIN users u2 ON u2.id = r.reported_id
		ORDER BY r.created_at DESC
		LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []ReportRow
	for rows.Next() {
		var row ReportRow
		var reporterEmail, reportedEmail pgtype.Text
		if err := rows.Scan(
			&row.ID, &row.ReporterID, &reporterEmail,
			&row.ReportedID, &reportedEmail,
			&row.Category, &row.Comment, &row.CreatedAt,
		); err != nil {
			return nil, err
		}
		if reporterEmail.Valid {
			row.ReporterEmail = reporterEmail.String
		}
		if reportedEmail.Valid {
			row.ReportedEmail = reportedEmail.String
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

// AdminAuditLog writes an immutable record of an admin action to admin_audit_log.
// targetType is a short label like "user", "message", "report".
// metadata is an optional map of additional context (target IDs, old values, etc.).
func (s *ModerationService) AdminAuditLog(ctx context.Context, adminID pgtype.UUID, action, targetType, targetID string, metadata map[string]any) {
	metaJSON, _ := json.Marshal(metadata)
	_, _ = s.DB.Exec(ctx, `
		INSERT INTO admin_audit_log(admin_id, action, target_type, target_id, metadata)
		VALUES ($1, $2, $3, $4, $5::jsonb)`,
		adminID, action, targetType, targetID, string(metaJSON))
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
