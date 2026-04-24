package clubs

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Gooowan/matchup/modules/core/types"
	gen "github.com/Gooowan/matchup/modules/clubs/gen"
)

type ClubService struct {
	DB      *pgxpool.Pool
	Queries *gen.Queries
}

func NewClubService(db *pgxpool.Pool) *ClubService {
	return &ClubService{DB: db, Queries: gen.New(db)}
}

var slugReplacer = regexp.MustCompile(`[^a-z0-9]+`)

func generateSlug(name string) string {
	s := strings.ToLower(name)
	s = slugReplacer.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}

type CreateClubParams struct {
	Name        string
	Description string
	Country     string
	City        string
	Address     string
	Latitude    float64
	Longitude   float64
	Website     string
	Phone       string
	IsVerified  bool
}

func (s *ClubService) CreateClub(ctx context.Context, p CreateClubParams) (gen.Club, error) {
	slug := generateSlug(p.Name)

	return s.Queries.CreateClub(ctx, gen.CreateClubParams{
		Name:        p.Name,
		Slug:        slug,
		Description: pgtype.Text{String: p.Description, Valid: p.Description != ""},
		Country:     p.Country,
		City:        p.City,
		Address:     pgtype.Text{String: p.Address, Valid: p.Address != ""},
		Latitude:    p.Latitude,
		Longitude:   p.Longitude,
		Website:     pgtype.Text{String: p.Website, Valid: p.Website != ""},
		Phone:       pgtype.Text{String: p.Phone, Valid: p.Phone != ""},
		IsVerified:  pgtype.Bool{Bool: p.IsVerified, Valid: true},
		Metadata:    types.JSONB{},
	})
}

func (s *ClubService) GetClubBySlug(ctx context.Context, slug string) (gen.Club, error) {
	return s.Queries.GetClubBySlug(ctx, slug)
}

func (s *ClubService) GetClubByID(ctx context.Context, id pgtype.UUID) (gen.Club, error) {
	return s.Queries.GetClubByID(ctx, id)
}

func (s *ClubService) ListClubs(ctx context.Context, country, city string, limit, offset int32) ([]gen.Club, error) {
	return s.Queries.ListClubs(ctx, gen.ListClubsParams{
		Country:   country,
		City:      city,
		LimitVal:  limit,
		OffsetVal: offset,
	})
}

func (s *ClubService) ListNearby(ctx context.Context, lat, lng float64, limit int32) ([]gen.ListClubsNearbyRow, error) {
	return s.Queries.ListClubsNearby(ctx, gen.ListClubsNearbyParams{
		Latitude:  lat,
		Longitude: lng,
		LimitVal:  limit,
	})
}

func (s *ClubService) AdminListClubs(ctx context.Context, limit, offset int32) ([]gen.Club, error) {
	return s.Queries.AdminListClubs(ctx, gen.AdminListClubsParams{
		LimitVal:  limit,
		OffsetVal: offset,
	})
}

func (s *ClubService) UpdateClub(ctx context.Context, id pgtype.UUID, p CreateClubParams) error {
	return s.Queries.UpdateClub(ctx, gen.UpdateClubParams{
		ID:          id,
		Name:        p.Name,
		Description: pgtype.Text{String: p.Description, Valid: p.Description != ""},
		Country:     p.Country,
		City:        p.City,
		Address:     pgtype.Text{String: p.Address, Valid: p.Address != ""},
		Latitude:    p.Latitude,
		Longitude:   p.Longitude,
		Website:     pgtype.Text{String: p.Website, Valid: p.Website != ""},
		Phone:       pgtype.Text{String: p.Phone, Valid: p.Phone != ""},
		Metadata:    types.JSONB{},
	})
}

func (s *ClubService) VerifyClub(ctx context.Context, id pgtype.UUID) error {
	return s.Queries.VerifyClub(ctx, id)
}

func (s *ClubService) DeactivateClub(ctx context.Context, id pgtype.UUID) error {
	return s.Queries.DeactivateClub(ctx, id)
}

func (s *ClubService) JoinClub(ctx context.Context, clubID, userID pgtype.UUID) error {
	return s.Queries.JoinClub(ctx, gen.JoinClubParams{
		ClubID: clubID,
		UserID: userID,
	})
}

func (s *ClubService) LeaveClub(ctx context.Context, clubID, userID pgtype.UUID) error {
	return s.Queries.LeaveClub(ctx, gen.LeaveClubParams{
		ClubID: clubID,
		UserID: userID,
	})
}

func (s *ClubService) GetUserClubs(ctx context.Context, userID pgtype.UUID) ([]gen.Club, error) {
	return s.Queries.GetUserClubs(ctx, userID)
}

func (s *ClubService) ListClubMembers(ctx context.Context, clubID pgtype.UUID, limit, offset int32) ([]gen.ListClubMembersRow, error) {
	return s.Queries.ListClubMembers(ctx, gen.ListClubMembersParams{
		ClubID:    clubID,
		LimitVal:  limit,
		OffsetVal: offset,
	})
}

func (s *ClubService) GetMemberCount(ctx context.Context, clubID pgtype.UUID) (int32, error) {
	return s.Queries.GetClubMemberCount(ctx, clubID)
}

func (s *ClubService) IsClubMember(ctx context.Context, clubID, userID pgtype.UUID) (bool, error) {
	return s.Queries.IsClubMember(ctx, gen.IsClubMemberParams{
		ClubID: clubID,
		UserID: userID,
	})
}

// RegisterClub creates an unverified club (for public self-registration form).
func (s *ClubService) RegisterClub(ctx context.Context, p CreateClubParams) (gen.Club, error) {
	if p.Country == "" || p.City == "" || p.Name == "" {
		return gen.Club{}, fmt.Errorf("name, country, and city are required")
	}
	p.IsVerified = false
	return s.CreateClub(ctx, p)
}

// ClaimClub sets the owner_user_id for a club if it has no owner yet.
// Returns the updated club or an error if already claimed.
func (s *ClubService) ClaimClub(ctx context.Context, clubID, userID pgtype.UUID) (gen.Club, error) {
	club, err := s.Queries.ClaimClub(ctx, gen.ClaimClubParams{
		ID:          clubID,
		OwnerUserID: userID,
	})
	if err != nil {
		return gen.Club{}, fmt.Errorf("already claimed or not found")
	}
	return club, nil
}

type ManageClubParams struct {
	Description  string
	Address      string
	Phone        string
	Website      string
	WorkingHours types.JSONB
}

// ManageClub lets the owner update their club's business details.
func (s *ClubService) ManageClub(ctx context.Context, clubID, ownerID pgtype.UUID, p ManageClubParams) error {
	return s.Queries.ManageClub(ctx, gen.ManageClubParams{
		ID:           clubID,
		OwnerUserID:  ownerID,
		Description:  pgtype.Text{String: p.Description, Valid: p.Description != ""},
		Address:      pgtype.Text{String: p.Address, Valid: p.Address != ""},
		Phone:        pgtype.Text{String: p.Phone, Valid: p.Phone != ""},
		Website:      pgtype.Text{String: p.Website, Valid: p.Website != ""},
		WorkingHours: p.WorkingHours,
	})
}

// ListOwnedClubs returns all active clubs owned by the given user.
func (s *ClubService) ListOwnedClubs(ctx context.Context, userID pgtype.UUID) ([]gen.Club, error) {
	return s.Queries.ListOwnedClubs(ctx, userID)
}
