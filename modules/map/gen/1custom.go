package gen

import (
	"github.com/Gooowan/matchup/modules/core/utils"
)

type LocationDTO struct {
	ID        string  `json:"id"`
	UserID    string  `json:"user_id"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	UpdatedAt int64   `json:"updated_at"`
}

func (l UserLocation) ToDTO() LocationDTO {
	return LocationDTO{
		ID:        utils.UUIDToString(l.ID),
		UserID:    utils.UUIDToString(l.UserID),
		Latitude:  l.Latitude,
		Longitude: l.Longitude,
		UpdatedAt: l.UpdatedAt.Time.UnixMilli(),
	}
}

type NearbyUserDTO struct {
	UserID     string  `json:"user_id"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	DistanceKm float64 `json:"distance_km"`
	UpdatedAt  int64   `json:"updated_at"`
}
