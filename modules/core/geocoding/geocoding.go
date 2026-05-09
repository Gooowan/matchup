// Package geocoding provides a server-side geocoding service backed by Nominatim.
// It applies a 1 rps rate limit (Nominatim policy), an in-memory LRU-ish cache,
// and falls back to a city centroid table when Nominatim cannot resolve an address.
package geocoding

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
	"unicode"
)

// ErrUnresolvable is returned when neither Nominatim nor the centroid table
// can produce coordinates for the given address.
var ErrUnresolvable = errors.New("geocoding: address could not be resolved")

// Geocoder resolves a human-readable address to WGS-84 coordinates.
type Geocoder interface {
	Geocode(ctx context.Context, country, city, address string) (lat, lng float64, err error)
}

// NominatimGeocoder implements Geocoder against the Nominatim search API.
type NominatimGeocoder struct {
	baseURL    string
	userAgent  string
	httpClient *http.Client

	mu      sync.Mutex
	cache   map[string][2]float64
	cacheOrd []string // tracks insertion order for eviction

	ticker *time.Ticker // 1 req/s rate limit
}

const maxCacheSize = 1024

// NewNominatimGeocoder creates a NominatimGeocoder.
// baseURL defaults to "https://nominatim.openstreetmap.org" and
// userAgent should identify your application as required by Nominatim policy.
func NewNominatimGeocoder(baseURL, userAgent string) *NominatimGeocoder {
	if baseURL == "" {
		baseURL = "https://nominatim.openstreetmap.org"
	}
	if userAgent == "" {
		userAgent = "matchup-server/1.0 (admin@matchup.local)"
	}
	return &NominatimGeocoder{
		baseURL:   strings.TrimRight(baseURL, "/"),
		userAgent: userAgent,
		httpClient: &http.Client{Timeout: 8 * time.Second},
		cache:      make(map[string][2]float64),
		ticker:     time.NewTicker(time.Second),
	}
}

// Close releases the internal rate-limit ticker.
func (g *NominatimGeocoder) Close() {
	g.ticker.Stop()
}

// Geocode resolves the address. When lat/lng are both 0 it falls back to the
// city centroid table. Returns ErrUnresolvable if nothing can be resolved.
func (g *NominatimGeocoder) Geocode(ctx context.Context, country, city, address string) (float64, float64, error) {
	normCountry := normalizeCountry(country)
	normCity := normalizeCity(city)
	query := buildQuery(address, normCity, normCountry)
	key := strings.ToLower(query)

	// Cache hit
	g.mu.Lock()
	if coords, ok := g.cache[key]; ok {
		g.mu.Unlock()
		return coords[0], coords[1], nil
	}
	g.mu.Unlock()

	// Rate-limit: wait for the next tick
	select {
	case <-g.ticker.C:
	case <-ctx.Done():
		return 0, 0, ctx.Err()
	}

	lat, lng, err := g.queryNominatim(ctx, query)
	if err == nil && (lat != 0 || lng != 0) {
		g.store(key, lat, lng)
		return lat, lng, nil
	}

	// Fallback: city centroid
	if c, ok := cityCentroid(normCountry, normCity); ok {
		g.store(key, c[0], c[1])
		return c[0], c[1], nil
	}

	return 0, 0, ErrUnresolvable
}

func (g *NominatimGeocoder) queryNominatim(ctx context.Context, query string) (float64, float64, error) {
	params := url.Values{}
	params.Set("q", query)
	params.Set("format", "json")
	params.Set("limit", "1")

	reqURL := fmt.Sprintf("%s/search?%s", g.baseURL, params.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return 0, 0, err
	}
	req.Header.Set("User-Agent", g.userAgent)
	req.Header.Set("Accept-Language", "en")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, 0, fmt.Errorf("nominatim returned %d", resp.StatusCode)
	}

	var results []struct {
		Lat string `json:"lat"`
		Lon string `json:"lon"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return 0, 0, err
	}
	if len(results) == 0 {
		return 0, 0, nil
	}

	var lat, lng float64
	fmt.Sscanf(results[0].Lat, "%f", &lat)
	fmt.Sscanf(results[0].Lon, "%f", &lng)
	return lat, lng, nil
}

func (g *NominatimGeocoder) store(key string, lat, lng float64) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if _, exists := g.cache[key]; !exists {
		if len(g.cacheOrd) >= maxCacheSize {
			oldest := g.cacheOrd[0]
			g.cacheOrd = g.cacheOrd[1:]
			delete(g.cache, oldest)
		}
		g.cacheOrd = append(g.cacheOrd, key)
	}
	g.cache[key] = [2]float64{lat, lng}
}

// buildQuery composes the Nominatim search string.
func buildQuery(address, city, country string) string {
	parts := []string{}
	if address != "" {
		parts = append(parts, address)
	}
	if city != "" {
		parts = append(parts, city)
	}
	if country != "" {
		parts = append(parts, country)
	}
	return strings.Join(parts, ", ")
}

// normalizeCountry maps common aliases to a canonical English country name.
func normalizeCountry(s string) string {
	s = strings.TrimSpace(s)
	switch strings.ToLower(s) {
	case "україна", "украина", "ukraine", "ua":
		return "ukraine"
	case "польща", "poland", "polska", "pl":
		return "poland"
	case "germany", "deutschland", "de":
		return "germany"
	default:
		return strings.ToLower(s)
	}
}

// cyrillicCityNames maps lowercase Ukrainian/Russian city names to their
// canonical English equivalents used by the centroid table and Nominatim.
var cyrillicCityNames = map[string]string{
	// Ukrainian spellings
	"київ":             "kyiv",
	"харків":           "kharkiv",
	"одеса":            "odesa",
	"дніпро":           "dnipro",
	"запоріжжя":        "zaporizhzhia",
	"львів":            "lviv",
	"кривий ріг":       "kryvyi rih",
	"миколаїв":         "mykolaiv",
	"маріуполь":        "mariupol",
	"луганськ":         "luhansk",
	"вінниця":          "vinnytsia",
	"херсон":           "kherson",
	"полтава":          "poltava",
	"чернігів":         "chernihiv",
	"черкаси":          "cherkasy",
	"суми":             "sumy",
	"житомир":          "zhytomyr",
	"хмельницький":     "khmelnytskyi",
	"рівне":            "rivne",
	"івано-франківськ": "ivano-frankivsk",
	"тернопіль":        "ternopil",
	"луцьк":            "lutsk",
	"ужгород":          "uzhhorod",
	"кропивницький":    "kropyvnytskyi",
	"чернівці":         "chernivtsi",
	// Russian spellings
	"киев":     "kyiv",
	"харьков":  "kharkiv",
	"одесса":   "odesa",
	"днепр":    "dnipro",
	"запорожье": "zaporizhzhia",
	"львов":    "lviv",
	"николаев": "mykolaiv",
	"чернигов": "chernihiv",
	"черкассы": "cherkasy",
	"луцк":     "lutsk",
}

// normalizeCity strips Ukrainian/Russian city prefixes, title-cases the result,
// and transliterates well-known Cyrillic city names to English so they match
// both the Nominatim query and the centroid lookup table.
func normalizeCity(s string) string {
	s = strings.TrimSpace(s)
	for _, prefix := range []string{"м.", "г.", "місто ", "город "} {
		if strings.HasPrefix(strings.ToLower(s), prefix) {
			s = s[len(prefix):]
			break
		}
	}
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	lower := strings.ToLower(s)
	if eng, ok := cyrillicCityNames[lower]; ok {
		return eng
	}
	return strings.ToLower(titleCase(s))
}

func titleCase(s string) string {
	runes := []rune(s)
	if len(runes) == 0 {
		return s
	}
	runes[0] = unicode.ToUpper(runes[0])
	for i := 1; i < len(runes); i++ {
		runes[i] = unicode.ToLower(runes[i])
	}
	return string(runes)
}
