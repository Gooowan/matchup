// cmd/seed-profiles creates mock dancer profiles for development and testing.
// It is idempotent: profiles whose email already exists are skipped.
//
// Usage:
//
//	DATABASE_URL=postgres://... go run ./cmd/seed-profiles [flags]
//
// Flags:
//
//	--count        number of profiles to create (default 30)
//	--assign-clubs randomly assign 70 % of profiles to an existing club (default true)
//	--password     bcrypt-hashed password for all seeded accounts (default "password123")
package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	count := flag.Int("count", 30, "number of profiles to seed")
	assignClubs := flag.Bool("assign-clubs", true, "randomly join 70% of profiles to an existing club")
	password := flag.String("password", "password123", "plain-text password to bcrypt for all seeded accounts")
	flag.Parse()

	ctx := context.Background()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		// Fall back to individual env vars similar to db.PostgresConnect
		dbURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
			getenv("POSTGRES_USER", "postgres"),
			getenv("POSTGRES_PASSWORD", "postgres"),
			getenv("POSTGRES_HOST", "localhost"),
			getenv("POSTGRES_PORT", "5432"),
			getenv("POSTGRES_DB", "matchup"),
			getenv("DB_SSL_MODE", "disable"),
		)
	}

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		fatalf("connect: %v", err)
	}
	defer pool.Close()

	hashBytes, err := bcrypt.GenerateFromPassword([]byte(*password), bcrypt.DefaultCost)
	if err != nil {
		fatalf("bcrypt: %v", err)
	}
	hash := string(hashBytes)

	// Load existing active club IDs
	var clubIDs []string
	if *assignClubs {
		rows, err := pool.Query(ctx, `SELECT id::text FROM clubs WHERE is_active = true`)
		if err != nil {
			fatalf("load clubs: %v", err)
		}
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err == nil {
				clubIDs = append(clubIDs, id)
			}
		}
		rows.Close()
		if len(clubIDs) == 0 {
			fmt.Println("⚠  No active clubs found; --assign-clubs will have no effect")
		}
	}

	created := 0
	skipped := 0

	for i := 1; i <= *count; i++ {
		email := fmt.Sprintf("seed-%03d@matchup.local", i)

		// Idempotency check
		var exists bool
		err := pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`, email).Scan(&exists)
		if err != nil {
			fatalf("check exists: %v", err)
		}
		if exists {
			skipped++
			continue
		}

		profile := randomProfile(i)

		// Insert user
		var userID string
		err = pool.QueryRow(ctx, `
			INSERT INTO users (email, password, role, profile_data)
			VALUES ($1, $2, 'USER', $3::jsonb)
			RETURNING id::text
		`, email, hash,
			fmt.Sprintf(`{"first_name":%q,"last_name":%q,"avatar":%q}`,
				profile.firstName, profile.lastName,
				fmt.Sprintf("https://i.pravatar.cc/600?u=%s", email)),
		).Scan(&userID)
		if err != nil {
			fatalf("insert user %d: %v", i, err)
		}

		// Insert profile
		_, err = pool.Exec(ctx, `
			INSERT INTO profiles
				(user_id, gender, birth_date, height_cm, goal, program, categories,
				 country, city, latitude, longitude, visible)
			VALUES ($1,$2,$3::date,$4,$5,$6,$7::text[],$8,$9,$10,$11,true)
		`, userID, profile.gender,
			profile.birthDate.Format("2006-01-02"),
			profile.heightCm,
			profile.goal,
			profile.program,
			pgTextArray(profile.categories),
			profile.country, profile.city,
			profile.lat, profile.lng,
		)
		if err != nil {
			fatalf("insert profile %d: %v", i, err)
		}

		// Insert preferences
		oppGender := "female"
		if profile.gender == "female" {
			oppGender = "male"
		}
		_, err = pool.Exec(ctx, `
			INSERT INTO user_preferences
				(user_id, preferred_gender, age_min, age_max, preferred_country, preferred_city)
			VALUES ($1,$2,$3,$4,$5,$6)
		`, userID, oppGender,
			profile.age()-5, profile.age()+5,
			profile.country, profile.city,
		)
		if err != nil {
			fatalf("insert preferences %d: %v", i, err)
		}

		// Optionally join a club (70 % chance)
		if *assignClubs && len(clubIDs) > 0 && rand.Float64() < 0.70 {
			clubID := clubIDs[rand.Intn(len(clubIDs))]
			_, _ = pool.Exec(ctx, `
				INSERT INTO club_members (club_id, user_id, role)
				VALUES ($1::uuid, $2::uuid, 'member')
				ON CONFLICT DO NOTHING
			`, clubID, userID)
		}

		created++
		fmt.Printf("  ✓ %s (%s %s)\n", email, profile.firstName, profile.lastName)
	}

	fmt.Printf("\nDone. Created %d profiles, skipped %d existing.\n", created, skipped)
}

// --- Data pools ---

var firstNamesMale = []string{
	"Олексій", "Дмитро", "Іван", "Максим", "Андрій",
	"Микола", "Сергій", "Владислав", "Артем", "Ігор",
	"Богдан", "Євген", "Роман", "Тарас", "Олег",
	"Михайло", "Павло", "Денис", "Юрій", "Віктор",
}

var firstNamesFemale = []string{
	"Олена", "Марія", "Тетяна", "Наталія", "Анна",
	"Оксана", "Вікторія", "Ірина", "Катерина", "Юлія",
	"Людмила", "Надія", "Галина", "Аліна", "Дарина",
	"Поліна", "Валентина", "Соломія", "Мар'яна", "Христина",
}

var lastNames = []string{
	"Шевченко", "Мельник", "Ковальчук", "Бондаренко", "Ткаченко",
	"Кравченко", "Олійник", "Шевчук", "Поліщук", "Гончаренко",
	"Іваненко", "Марченко", "Коваленко", "Романенко", "Савченко",
	"Ткачук", "Петренко", "Мороз", "Лисенко", "Давиденко",
}

var categories = []string{"salsa", "bachata", "kizomba", "tango", "waltz", "hip-hop", "contemporary", "latin"}
var goals = []string{"hobby", "sport", "professional"}
var programs = []string{"standard", "sport"}

type cityEntry struct {
	city string
	lat  float64
	lng  float64
}

var ukraineCities = []cityEntry{
	{"Kyiv", 50.4501, 30.5234},
	{"Kharkiv", 49.9935, 36.2304},
	{"Lviv", 49.8397, 24.0297},
	{"Odesa", 46.4825, 30.7233},
	{"Dnipro", 48.4647, 35.0462},
	{"Zaporizhzhia", 47.8388, 35.1396},
	{"Vinnytsia", 49.2331, 28.4682},
	{"Poltava", 49.5883, 34.5514},
	{"Chernihiv", 51.4982, 31.2893},
	{"Ivano-Frankivsk", 48.9226, 24.7111},
}

type profileData struct {
	firstName  string
	lastName   string
	gender     string
	birthDate  time.Time
	heightCm   int
	goal       string
	program    string
	categories []string
	country    string
	city       string
	lat        float64
	lng        float64
}

func (p profileData) age() int {
	return int(time.Since(p.birthDate).Hours() / 8766)
}

func randomProfile(seed int) profileData {
	r := rand.New(rand.NewSource(int64(seed) * 1337))

	gender := "male"
	firstName := firstNamesMale[r.Intn(len(firstNamesMale))]
	if r.Intn(2) == 0 {
		gender = "female"
		firstName = firstNamesFemale[r.Intn(len(firstNamesFemale))]
	}

	age := 18 + r.Intn(22) // 18–39
	birthYear := time.Now().Year() - age
	birthDate := time.Date(birthYear, time.Month(1+r.Intn(12)), 1+r.Intn(28), 0, 0, 0, 0, time.UTC)

	nc := 1 + r.Intn(3)
	cats := make([]string, nc)
	used := map[int]bool{}
	for j := 0; j < nc; j++ {
		for {
			idx := r.Intn(len(categories))
			if !used[idx] {
				used[idx] = true
				cats[j] = categories[idx]
				break
			}
		}
	}

	city := ukraineCities[r.Intn(len(ukraineCities))]

	// Small jitter so profiles don't overlap exactly on city centroid
	jitter := func() float64 { return (r.Float64() - 0.5) * 0.05 }

	return profileData{
		firstName:  firstName,
		lastName:   lastNames[r.Intn(len(lastNames))],
		gender:     gender,
		birthDate:  birthDate,
		heightCm:   155 + r.Intn(35),
		goal:       goals[r.Intn(len(goals))],
		program:    programs[r.Intn(len(programs))],
		categories: cats,
		country:    "Ukraine",
		city:       city.city,
		lat:        city.lat + jitter(),
		lng:        city.lng + jitter(),
	}
}

// pgTextArray formats a Go string slice as PostgreSQL array literal e.g. {"salsa","bachata"}
func pgTextArray(ss []string) string {
	if len(ss) == 0 {
		return "{}"
	}
	s := "{"
	for i, v := range ss {
		if i > 0 {
			s += ","
		}
		s += `"` + v + `"`
	}
	return s + "}"
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "seed-profiles: "+format+"\n", args...)
	os.Exit(1)
}
