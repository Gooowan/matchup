// Package gmaps provides a Google Places API client for resolving club data
// from Google Maps URLs. When GOOGLE_PLACES_API_KEY is empty the client is
// a no-op stub that returns ErrNotConfigured.
package gmaps

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ErrNotConfigured is returned when no API key has been provided.
var ErrNotConfigured = errors.New("google places: GOOGLE_PLACES_API_KEY not configured")

// ErrQuotaExceeded is returned when the daily LookupURL cap has been reached.
var ErrQuotaExceeded = errors.New("google places: daily quota exceeded")

// dailyQuota is a mutex-guarded counter that resets at midnight UTC.
type dailyQuota struct {
	mu      sync.Mutex
	count   int
	limit   int
	resetAt time.Time
}

func (q *dailyQuota) allow() bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	now := time.Now().UTC()
	if now.After(q.resetAt) {
		y, m, d := now.Date()
		q.resetAt = time.Date(y, m, d+1, 0, 0, 0, 0, time.UTC)
		q.count = 0
	}
	if q.count >= q.limit {
		return false
	}
	q.count++
	return true
}

// PlaceData holds the club details extracted from a Google Maps link.
type PlaceData struct {
	Name         string            `json:"name"`
	Address      string            `json:"address"`
	City         string            `json:"city"`
	Country      string            `json:"country"`
	Latitude     float64           `json:"latitude"`
	Longitude    float64           `json:"longitude"`
	Website      string            `json:"website"`
	Phone        string            `json:"phone"`
	WorkingHours map[string]string `json:"working_hours,omitempty"` // {"Monday":"9:00 AM – 9:00 PM",...}
	Photos       []string          `json:"photos"`                  // proxy URLs: /images/gphoto?ref=...&w=800
}

// Client resolves a Google Maps URL to PlaceData.
type Client interface {
	LookupURL(ctx context.Context, rawURL string) (*PlaceData, error)
}

// GooglePlacesClient is the production implementation.
type GooglePlacesClient struct {
	apiKey     string
	httpClient *http.Client
	quota      *dailyQuota
}

// NewGooglePlacesClient creates a client. When apiKey is empty every call
// returns ErrNotConfigured so local dev without a key still compiles & runs.
// The daily lookup cap defaults to 50 and can be overridden via GOOGLE_PLACES_DAILY_LIMIT.
func NewGooglePlacesClient(apiKey string) *GooglePlacesClient {
	limit := 50
	if v := os.Getenv("GOOGLE_PLACES_DAILY_LIMIT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	now := time.Now().UTC()
	y, m, d := now.Date()
	return &GooglePlacesClient{
		apiKey: apiKey,
		quota: &dailyQuota{
			limit:   limit,
			resetAt: time.Date(y, m, d+1, 0, 0, 0, 0, time.UTC),
		},
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse // let us follow manually
			},
		},
	}
}

// IsConfigured reports whether an API key is set.
func (c *GooglePlacesClient) IsConfigured() bool {
	return c.apiKey != ""
}

// APIKey returns the configured API key for use by internal proxy handlers.
func (c *GooglePlacesClient) APIKey() string {
	return c.apiKey
}

// LookupURL resolves a raw Google Maps URL (short or long) to PlaceData.
func (c *GooglePlacesClient) LookupURL(ctx context.Context, rawURL string) (*PlaceData, error) {
	if !c.IsConfigured() {
		return nil, ErrNotConfigured
	}
	if !c.quota.allow() {
		return nil, ErrQuotaExceeded
	}

	longURL, err := c.resolveURL(ctx, rawURL)
	if err != nil {
		return nil, fmt.Errorf("resolve url: %w", err)
	}

	placeID, err := extractPlaceID(longURL)
	if err != nil || placeID == "" {
		// Fall back: try findplacefromtext using coords from URL
		lat, lng := extractLatLng(longURL)
		name := extractNameHint(longURL)
		if name == "" && lat == 0 {
			return nil, fmt.Errorf("could not determine place from URL")
		}
		placeID, err = c.findPlaceID(ctx, name, lat, lng)
		if err != nil {
			return nil, fmt.Errorf("find place id: %w", err)
		}
	}

	details, err := c.placeDetails(ctx, placeID)
	if err != nil {
		return nil, fmt.Errorf("place details: %w", err)
	}

	photos := proxyPhotoURLs(details.photoRefs)

	return &PlaceData{
		Name:         details.name,
		Address:      details.address,
		City:         details.city,
		Country:      details.country,
		Latitude:     details.lat,
		Longitude:    details.lng,
		Website:      details.website,
		Phone:        details.phone,
		WorkingHours: parseWeekdayText(details.weekdayText),
		Photos:       photos,
	}, nil
}

// allowedMapsHosts is the strict allowlist of hosts that LookupURL may contact.
// Any other host (including internal addresses and cloud metadata endpoints) is
// rejected before the first outbound request and after every redirect hop.
var allowedMapsHosts = []string{
	"google.com",
	"maps.google.com",
	"www.google.com",
	"maps.app.goo.gl",
	"goo.gl",
	"maps.googleapis.com",
	"places.googleapis.com",
}

// isAllowedMapsHost returns true when host (with port stripped) matches an
// entry in allowedMapsHosts exactly or is a subdomain of google.com.
func isAllowedMapsHost(host string) bool {
	// Strip port if present.
	if h, _, found := strings.Cut(host, ":"); found {
		host = h
	}
	for _, allowed := range allowedMapsHosts {
		if host == allowed {
			return true
		}
		// Allow any *.google.com subdomain (e.g. maps.google.com already listed,
		// but future subdomains like lh3.googleusercontent.com are intentionally
		// NOT included — keep the list explicit).
	}
	return false
}

// resolveURL follows HTTP redirects to expand short Maps URLs.
// Every URL — including the initial one and every redirect destination — is
// validated against allowedMapsHosts before a request is made, preventing SSRF
// via user-supplied URLs or redirect chains that lead to internal addresses.
func (c *GooglePlacesClient) resolveURL(ctx context.Context, rawURL string) (string, error) {
	current := rawURL
	for i := 0; i < 5; i++ {
		parsed, err := url.Parse(current)
		if err != nil {
			return "", fmt.Errorf("invalid url: %w", err)
		}
		if !isAllowedMapsHost(parsed.Host) {
			return "", fmt.Errorf("ssrf-guard: host %q is not in the Google Maps allowlist", parsed.Host)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, current, nil)
		if err != nil {
			return "", err
		}
		req.Header.Set("User-Agent", "Mozilla/5.0")
		resp, err := c.httpClient.Do(req)
		if err != nil {
			return "", err
		}
		resp.Body.Close()

		loc := resp.Header.Get("Location")
		if loc == "" {
			return current, nil
		}
		// Resolve relative redirects
		base, _ := url.Parse(current)
		next, err := base.Parse(loc)
		if err != nil {
			return current, nil
		}
		current = next.String()
	}
	return current, nil
}

// extractPlaceID tries to pull a place_id from the Maps URL.
// Modern Maps URLs embed it as the last path segment of /place/... or as a query param.
func extractPlaceID(longURL string) (string, error) {
	u, err := url.Parse(longURL)
	if err != nil {
		return "", err
	}
	// Check query param
	if pid := u.Query().Get("place_id"); pid != "" {
		return pid, nil
	}
	// /maps/place/.../data=...!1s<placeID>!...
	// The place ID in modern URLs starts with ChIJ
	if idx := strings.Index(longURL, "!1s"); idx != -1 {
		rest := longURL[idx+3:]
		if end := strings.IndexAny(rest, "!&"); end != -1 {
			rest = rest[:end]
		}
		pid := strings.TrimSpace(rest)
		if strings.HasPrefix(pid, "ChIJ") && len(pid) > 10 {
			return pid, nil
		}
	}
	return "", nil
}

func extractLatLng(longURL string) (float64, float64) {
	// @lat,lng,zoom pattern
	u, _ := url.Parse(longURL)
	parts := strings.Split(u.Path, "@")
	if len(parts) < 2 {
		return 0, 0
	}
	coords := strings.SplitN(parts[1], ",", 3)
	if len(coords) < 2 {
		return 0, 0
	}
	var lat, lng float64
	fmt.Sscanf(coords[0], "%f", &lat)
	fmt.Sscanf(coords[1], "%f", &lng)
	return lat, lng
}

func extractNameHint(longURL string) string {
	u, _ := url.Parse(longURL)
	// /maps/place/<name>/...
	parts := strings.Split(u.Path, "/")
	for i, p := range parts {
		if p == "place" && i+1 < len(parts) {
			name, _ := url.PathUnescape(parts[i+1])
			return name
		}
	}
	return ""
}

// findPlaceID calls the Places findplacefromtext endpoint.
func (c *GooglePlacesClient) findPlaceID(ctx context.Context, name string, lat, lng float64) (string, error) {
	input := name
	if input == "" {
		input = fmt.Sprintf("%f,%f", lat, lng)
	}
	params := url.Values{}
	params.Set("input", input)
	params.Set("inputtype", "textquery")
	params.Set("fields", "place_id")
	params.Set("key", c.apiKey)
	if lat != 0 {
		params.Set("locationbias", fmt.Sprintf("circle:500@%f,%f", lat, lng))
	}

	reqURL := "https://maps.googleapis.com/maps/api/place/findplacefromtext/json?" + params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return "", err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Candidates []struct {
			PlaceID string `json:"place_id"`
		} `json:"candidates"`
		Status string `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if result.Status != "OK" && result.Status != "ZERO_RESULTS" {
		return "", fmt.Errorf("places api: %s", result.Status)
	}
	if len(result.Candidates) == 0 {
		return "", fmt.Errorf("no place found")
	}
	return result.Candidates[0].PlaceID, nil
}

type placeDetails struct {
	name         string
	address      string
	city         string
	country      string
	lat, lng     float64
	website      string
	phone        string
	weekdayText  []string // raw ["Monday: 9:00 AM – 9:00 PM", ...]
	photoRefs    []string
}

// placeDetails calls the Place Details API.
func (c *GooglePlacesClient) placeDetails(ctx context.Context, placeID string) (*placeDetails, error) {
	params := url.Values{}
	params.Set("place_id", placeID)
	params.Set("fields", "name,formatted_address,geometry,address_components,photos,website,international_phone_number,opening_hours")
	params.Set("key", c.apiKey)

	reqURL := "https://maps.googleapis.com/maps/api/place/details/json?" + params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Result struct {
			Name             string `json:"name"`
			FormattedAddress string `json:"formatted_address"`
			Geometry         struct {
				Location struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				} `json:"location"`
			} `json:"geometry"`
			AddressComponents []struct {
				LongName  string   `json:"long_name"`
				ShortName string   `json:"short_name"`
				Types     []string `json:"types"`
			} `json:"address_components"`
			Photos []struct {
				PhotoReference string `json:"photo_reference"`
			} `json:"photos"`
			Website                  string `json:"website"`
			InternationalPhoneNumber string `json:"international_phone_number"`
			OpeningHours             struct {
				WeekdayText []string `json:"weekday_text"`
			} `json:"opening_hours"`
		} `json:"result"`
		Status string `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if result.Status != "OK" {
		return nil, fmt.Errorf("place details api: %s", result.Status)
	}

	d := &placeDetails{
		name:        result.Result.Name,
		address:     result.Result.FormattedAddress,
		lat:         result.Result.Geometry.Location.Lat,
		lng:         result.Result.Geometry.Location.Lng,
		website:     result.Result.Website,
		phone:       result.Result.InternationalPhoneNumber,
		weekdayText: result.Result.OpeningHours.WeekdayText,
	}

	for _, comp := range result.Result.AddressComponents {
		for _, t := range comp.Types {
			if t == "locality" {
				d.city = comp.LongName
			}
			if t == "country" {
				d.country = comp.LongName
			}
		}
	}

	// Cap photos at 5
	max := len(result.Result.Photos)
	if max > 5 {
		max = 5
	}
	for i := 0; i < max; i++ {
		d.photoRefs = append(d.photoRefs, result.Result.Photos[i].PhotoReference)
	}

	return d, nil
}

// proxyPhotoURLs converts Google photo_reference strings into backend proxy URLs.
// No network call is made; the actual photo bytes are fetched on demand via
// GET /images/gphoto. Capped at 5 photos.
func proxyPhotoURLs(refs []string) []string {
	if len(refs) == 0 {
		return nil
	}
	max := len(refs)
	if max > 5 {
		max = 5
	}
	urls := make([]string, 0, max)
	for i := 0; i < max; i++ {
		urls = append(urls, "/images/gphoto?ref="+url.QueryEscape(refs[i])+"&w=800")
	}
	return urls
}

// parseWeekdayText converts the Places API weekday_text array into a simple
// day→hours map suitable for the clubs.working_hours JSONB column.
// Input:  ["Monday: 9:00 AM – 9:00 PM", "Tuesday: Closed", ...]
// Output: {"Monday":"9:00 AM – 9:00 PM","Tuesday":"Closed", ...}
func parseWeekdayText(lines []string) map[string]string {
	if len(lines) == 0 {
		return nil
	}
	out := make(map[string]string, len(lines))
	for _, line := range lines {
		idx := strings.Index(line, ": ")
		if idx < 0 {
			continue
		}
		day := strings.TrimSpace(line[:idx])
		hours := strings.TrimSpace(line[idx+2:])
		if day != "" && hours != "" {
			out[day] = hours
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
