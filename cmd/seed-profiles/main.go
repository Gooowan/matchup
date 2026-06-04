// cmd/seed-profiles creates mock dancer profiles for development and testing.
// It is idempotent: profiles whose email already exists are skipped.
//
// Usage:
//
//	DATABASE_URL=postgres://... go run ./cmd/seed-profiles [flags]
//
// Flags:
//
//	--count        number of profiles to create (default 70)
//	--gender-split balanced|female|male  (default "balanced")
//	--assign-clubs randomly assign 70 % of profiles to an existing club (default true)
//	--password     bcrypt-hashed password for all seeded accounts (default "password123")
//	--no-minio     skip MinIO upload; use external placeholder URLs instead
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	count := flag.Int("count", 70, "number of profiles to seed")
	genderSplit := flag.String("gender-split", "balanced", "gender distribution: balanced|female|male")
	assignClubs := flag.Bool("assign-clubs", true, "randomly join 70% of profiles to an existing club")
	password := flag.String("password", "password123", "plain-text password to bcrypt for all seeded accounts")
	noMinio := flag.Bool("no-minio", false, "skip MinIO upload; use external placeholder URLs")
	trainerCount := flag.Int("trainers", 12, "number of trainer profiles to seed (0 to skip)")
	flag.Parse()

	ctx := context.Background()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
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

	// Optional MinIO client
	var mc *minio.Client
	var publicEndpoint string
	if !*noMinio {
		mc, publicEndpoint, err = newMinioClient()
		if err != nil {
			fmt.Printf("⚠  MinIO unavailable (%v); falling back to external URLs\n", err)
			mc = nil
		} else {
			ensureBuckets(ctx, mc)
		}
	}

	// v1: all profiles are Kyiv-based. Load only Kyiv clubs so coords stay in the city.
	type clubEntry struct {
		id  string
		lat float64
		lng float64
	}
	var kyivClubEntries []clubEntry
	{
		rows, qErr := pool.Query(ctx, `
			SELECT id::text, latitude, longitude
			FROM clubs
			WHERE is_active = true AND city IN ('Київ','Kyiv')
			  AND latitude != 0 AND longitude != 0`)
		if qErr != nil {
			fatalf("load kyiv clubs: %v", qErr)
		}
		for rows.Next() {
			var e clubEntry
			if scanErr := rows.Scan(&e.id, &e.lat, &e.lng); scanErr == nil {
				kyivClubEntries = append(kyivClubEntries, e)
			}
		}
		rows.Close()
		if len(kyivClubEntries) == 0 {
			fatalf("no active Kyiv clubs found — run the app and create at least one first")
		}
		fmt.Printf("ℹ  Found %d Kyiv clubs to assign\n", len(kyivClubEntries))
	}
	// Keep the old clubIDs variable unused; we now only ever use kyivClubEntries.
	_ = assignClubs

	created := 0
	skipped := 0

	for i := 1; i <= *count; i++ {
		email := fmt.Sprintf("seed-%03d@matchup.local", i)

		// Idempotency check
		var exists bool
		if chkErr := pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`, email).Scan(&exists); chkErr != nil {
			fatalf("check exists: %v", chkErr)
		}
		if exists {
			skipped++
			continue
		}

		profile := randomProfile(i, *genderSplit)

		// --- Resolve photo URLs (upload to MinIO or use external) ---
		avatarURL, galleryURLs := resolvePhotos(ctx, mc, publicEndpoint, email, profile, i)

		// profile_data for users table
		profileDataJSON, _ := json.Marshal(map[string]any{
			"first_name": profile.firstName,
			"last_name":  profile.lastName,
			"avatar":     avatarURL,
		})

		// Insert user
		var userID string
		if insErr := pool.QueryRow(ctx, `
			INSERT INTO users (email, password, role, profile_data)
			VALUES ($1, $2, 'USER', $3::jsonb)
			RETURNING id::text
		`, email, hash, string(profileDataJSON)).Scan(&userID); insErr != nil {
			fatalf("insert user %d: %v", i, insErr)
		}

		// metadata (bio + media_urls)
		metaJSON, _ := json.Marshal(map[string]any{
			"bio":        profile.bio,
			"media_urls": galleryURLs,
		})

		// Always assign a Kyiv club; coords are locked to the club.
		assignedClub := kyivClubEntries[rand.Intn(len(kyivClubEntries))]
		var clubID string
		if insErr2 := pool.QueryRow(ctx, `
			INSERT INTO club_members (club_id, user_id, role)
			VALUES ($1::uuid, $2::uuid, 'member')
			ON CONFLICT DO NOTHING
			RETURNING club_id::text
		`, assignedClub.id, userID).Scan(&clubID); insErr2 != nil {
			clubID = assignedClub.id
		}

		// Insert profile — coords locked to the club, city/country = Київ/Україна.
		_, insErr := pool.Exec(ctx, `
			INSERT INTO profiles
				(user_id, gender, birth_date, height_cm, goal, program,
				 dance_styles, categories, country, city, latitude, longitude,
				 visible, primary_club_id, metadata)
			VALUES ($1,$2,$3::date,$4,$5,$6,$7::text[],$8::text[],$9,$10,$11,$12,true,$13::uuid,$14::jsonb)
		`, userID, profile.gender,
			profile.birthDate.Format("2006-01-02"),
			profile.heightCm,
			profile.goal,
			profile.program,
			pgTextArray(profile.danceStyles),
			pgTextArray(profile.categories),
			"Україна", "Київ",
			assignedClub.lat, assignedClub.lng,
			assignedClub.id,
			string(metaJSON),
		)
		if insErr != nil {
			fatalf("insert profile %d: %v", i, insErr)
		}

		// Insert preferences — prefer opposite gender, age ±7, locked to Київ/Україна.
		oppGender := "female"
		if profile.gender == "female" {
			oppGender = "male"
		}
		_, insErr = pool.Exec(ctx, `
			INSERT INTO user_preferences
				(user_id, preferred_gender, age_min, age_max, preferred_country, preferred_city)
			VALUES ($1,$2,$3,$4,$5,$6)
		`, userID, oppGender,
			profile.age()-7, profile.age()+7,
			"Україна", "Київ",
		)
		if insErr != nil {
			fatalf("insert preferences %d: %v", i, insErr)
		}

		created++
		fmt.Printf("  ✓ %s (%s %s, %s)\n", email, profile.firstName, profile.lastName, profile.gender)
	}

	fmt.Printf("\nDone. Created %d profiles, skipped %d existing.\n", created, skipped)

	// --- Trainer seeding ---
	if *trainerCount > 0 {
		fmt.Printf("\n=== Seeding %d trainer profiles ===\n", *trainerCount)
		trainerNames := [][2]string{
			{"Oleksii", "Kovalenko"}, {"Maria", "Savchenko"}, {"Dmytro", "Petrenko"},
			{"Iryna", "Shevchenko"}, {"Mykola", "Bondarenko"}, {"Yulia", "Moroz"},
			{"Andriy", "Tkachenko"}, {"Oksana", "Lysenko"}, {"Vladyslav", "Kravchenko"},
			{"Tetiana", "Ilchenko"}, {"Roman", "Sydorenko"}, {"Natalia", "Karpenko"},
		}
		trainerBios := []string{
			"Professional Latin dance coach, 10 years experience. WDC certified judge.",
			"Ballroom specialist with European championship titles. Training couples of all levels.",
			"Contemporary and street dance trainer. Former Broadway performer.",
			"Salsa & Bachata instructor, teaching private and group classes.",
			"Classical ballet background, now specializing in competitive ballroom.",
			"Argentine Tango expert. Teaching in Kyiv and internationally.",
			"Specialized in kids and teen dance education. Fun and professional.",
			"Hip-hop and breaking coach with a competition background.",
			"Bolero and Pasodoble master. WDSF Licensed instructor.",
			"Swing and Jive coach, bringing the joy of classic dances to life.",
			"Modern jazz and contemporary teacher. Choreo for stage and competition.",
			"Certified Zumba instructor and fitness dance coach.",
		}
		trainerStyles := [][]string{
			{"Latin", "Cha-cha-cha"}, {"Ballroom", "Waltz"}, {"Contemporary", "Hip-hop"},
			{"Salsa", "Bachata"}, {"Ballroom", "Latin"}, {"Argentine Tango"},
			{"Kids Dance", "Ballet"}, {"Hip-hop", "Breaking"}, {"Standard", "Latin"},
			{"Swing", "Jive"}, {"Jazz", "Contemporary"}, {"Zumba", "Latin"},
		}
		genders := []string{"male", "female", "male", "female", "male", "female",
			"female", "male", "male", "female", "female", "female"}

		tCreated, tSkipped := 0, 0
		for i := 0; i < *trainerCount && i < len(trainerNames); i++ {
			email := fmt.Sprintf("trainer-%02d@matchup.local", i+1)
			var exists bool
			if chkErr := pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`, email).Scan(&exists); chkErr != nil {
				fatalf("check trainer exists: %v", chkErr)
			}
			if exists {
				tSkipped++
				continue
			}

			nm := trainerNames[i]
			gender := genders[i]
			styles := trainerStyles[i]
			bio := trainerBios[i]

			pdJSON, _ := json.Marshal(map[string]any{
				"first_name":   nm[0],
				"last_name":    nm[1],
				"account_type": "trainer",
				"avatar":       fmt.Sprintf("https://images.unsplash.com/photo-1526413232644-8a40f03cc03b?w=600&q=80&auto=format&fit=crop&seed=%d", i),
			})
			metaJSON, _ := json.Marshal(map[string]any{
				"bio":          bio,
				"account_type": "trainer",
				"role":         "trainer",
			})

			var trainerUserID string
			if insErr := pool.QueryRow(ctx, `
				INSERT INTO users (email, password, role, profile_data)
				VALUES ($1, $2, 'USER', $3::jsonb)
				RETURNING id::text
			`, email, hash, string(pdJSON)).Scan(&trainerUserID); insErr != nil {
				fatalf("insert trainer user %d: %v", i, insErr)
			}

			// Assign to a club
			assignedClub := kyivClubEntries[rand.Intn(len(kyivClubEntries))]

			_, insErr := pool.Exec(ctx, `
				INSERT INTO profiles
					(user_id, account_type, gender, birth_date, height_cm, goal, program,
					 dance_styles, categories, country, city, latitude, longitude,
					 visible, primary_club_id, metadata)
				VALUES ($1,'trainer',$2,$3::date,$4,'professional','standard',
					$5::text[],$6::text[],'Україна','Київ',$7,$8,true,$9::uuid,$10::jsonb)
			`, trainerUserID, gender,
				"1985-06-15",
				170,
				pgTextArray(styles),
				pgTextArray(styles),
				assignedClub.lat, assignedClub.lng,
				assignedClub.id,
				string(metaJSON),
			)
			if insErr != nil {
				fatalf("insert trainer profile %d: %v", i, insErr)
			}

			// Register in club_trainers junction table.
			_, _ = pool.Exec(ctx, `
				INSERT INTO club_trainers (club_id, trainer_user_id)
				VALUES ($1::uuid, $2::uuid)
				ON CONFLICT DO NOTHING
			`, assignedClub.id, trainerUserID)

			tCreated++
			fmt.Printf("  ✓ %s (%s %s, trainer)\n", email, nm[0], nm[1])
		}
		fmt.Printf("\nTrainers: Created %d, skipped %d existing.\n", tCreated, tSkipped)
	}
}

// resolvePhotos downloads images and uploads them to MinIO, returning the public URLs.
// Falls back to external placeholder URLs when mc is nil.
func resolvePhotos(
	ctx context.Context,
	mc *minio.Client,
	publicEndpoint string,
	email string,
	profile profileData,
	seed int,
) (avatarURL string, galleryURLs []string) {
	r := rand.New(rand.NewSource(int64(seed) * 7919))

	// Pick deterministic Unsplash photo IDs from our curated pools.
	// We use separate pools for male/female to keep photos gender-relevant.
	pool := unsplashMale
	if profile.gender == "female" {
		pool = unsplashFemale
	}
	photoCount := 2 + r.Intn(3) // 2-4 gallery photos
	if photoCount > len(pool)-1 {
		photoCount = len(pool) - 1
	}

	avatarPhotoID := pool[seed%len(pool)]
	galleryPhotoIDs := make([]string, photoCount)
	for j := 0; j < photoCount; j++ {
		galleryPhotoIDs[j] = pool[(seed+j+1)%len(pool)]
	}

	// External source URLs
	avatarSrc := fmt.Sprintf("https://images.unsplash.com/photo-%s?w=600&q=80&auto=format&fit=crop", avatarPhotoID)
	gallerySrcs := make([]string, photoCount)
	for j, pid := range galleryPhotoIDs {
		gallerySrcs[j] = fmt.Sprintf("https://images.unsplash.com/photo-%s?w=600&q=80&auto=format&fit=crop", pid)
	}

	if mc == nil {
		// Use external URLs directly (no MinIO)
		avatarURL = avatarSrc
		galleryURLs = gallerySrcs
		return
	}

	ts := time.Now().UnixMilli()

	// Upload avatar
	avatarKey := fmt.Sprintf("%s_%d.jpg", sanitizeEmail(email), ts)
	if uploadErr := uploadFromURL(ctx, mc, "avatars", avatarKey, avatarSrc); uploadErr != nil {
		fmt.Printf("    ⚠ avatar upload failed (%v), using external URL\n", uploadErr)
		avatarURL = avatarSrc
	} else {
		avatarURL = fmt.Sprintf("%s/avatars/%s", strings.TrimRight(publicEndpoint, "/"), avatarKey)
	}

	// Upload gallery photos
	galleryURLs = make([]string, photoCount)
	for j, src := range gallerySrcs {
		photoKey := fmt.Sprintf("%s_%d_%d.jpg", sanitizeEmail(email), ts, j)
		if uploadErr := uploadFromURL(ctx, mc, "photos", photoKey, src); uploadErr != nil {
			fmt.Printf("    ⚠ gallery photo %d upload failed (%v), using external URL\n", j, uploadErr)
			galleryURLs[j] = src
		} else {
			galleryURLs[j] = fmt.Sprintf("%s/photos/%s", strings.TrimRight(publicEndpoint, "/"), photoKey)
		}
	}
	return
}

// uploadFromURL downloads src and puts it into bucket/key on MinIO.
func uploadFromURL(ctx context.Context, mc *minio.Client, bucket, key, src string) error {
	resp, err := http.Get(src) //nolint:noctx
	if err != nil {
		return fmt.Errorf("download: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download: HTTP %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read body: %w", err)
	}

	_, err = mc.PutObject(ctx, bucket, key, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{
		ContentType: "image/jpeg",
	})
	return err
}

func newMinioClient() (*minio.Client, string, error) {
	endpoint := getenv("MINIO_ENDPOINT", "localhost:9000")
	access := getenv("MINIO_ACCESS_KEY", "minioadmin")
	secret := getenv("MINIO_SECRET_KEY", "minioadmin")
	publicEP := getenv("MINIO_PUBLIC_ENDPOINT", "http://localhost:9000")

	// Strip protocol from endpoint for the minio client
	cleanEndpoint := strings.TrimPrefix(strings.TrimPrefix(endpoint, "https://"), "http://")
	useSSL := strings.HasPrefix(endpoint, "https://")

	mc, err := minio.New(cleanEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(access, secret, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, "", err
	}

	// Quick connectivity check
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if _, err := mc.ListBuckets(ctx); err != nil {
		return nil, "", fmt.Errorf("ping: %w", err)
	}

	return mc, publicEP, nil
}

func ensureBuckets(ctx context.Context, mc *minio.Client) {
	publicBuckets := []string{"avatars", "photos"}
	for _, b := range publicBuckets {
		exists, _ := mc.BucketExists(ctx, b)
		if !exists {
			_ = mc.MakeBucket(ctx, b, minio.MakeBucketOptions{})
			policy := fmt.Sprintf(`{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":"s3:GetObject","Resource":"arn:aws:s3:::%s/*"}]}`, b)
			_ = mc.SetBucketPolicy(ctx, b, policy)
		}
	}
}

func sanitizeEmail(email string) string {
	return strings.NewReplacer("@", "_", ".", "_").Replace(email)
}

// --- Curated Unsplash photo IDs (free-use, deterministic) ---
// Format: numeric Unsplash photo ID (used in image URL)

var unsplashMale = []string{
	"1507003211169-0a1dd7228f2d", "1506794778202-cad84cf45f1d", "1500648767791-00dcc994a43e",
	"1472099645785-5658abf4ff4e", "1519085360753-af0119f7cbe7", "1558618666-fcd25c85cd64",
	"1492562080023-ab3db95bfbce", "1531746020798-e6953c6e8e04", "1544723795-3fb6469f5b39",
	"1488161628813-04466f872be2", "1463453091185-61582044d556", "1478060780-e2e7c64f7cb3",
	"1504257432389-52343af06ae3", "1532170579297-281918bf2807", "1499996755657-11b093c48c85",
	"1521119989659-3ec900e3a10f", "1557862921-37829c790f19", "1570295999919-56ceb5ecca61",
	"1552642762-f55d06580015", "1545167622-43594203f468",
}

var unsplashFemale = []string{
	"1494790108377-be9c29b29330", "1529626455594-4ff0802cfb7e", "1524504388881-64cd788fa17c",
	"1488426862026-3ee34a7d66df", "1487412720507-e7ab37603c6f", "1508214751196-bcfd4ca60f91",
	"1531123897727-240f11e1bb78", "1488207272568-b0d3f5a84800", "1534528741775-53994a69daeb",
	"1520813792240-56fc4a3765a7", "1517841905240-472988babdf9", "1502003148287-a8ef8c5e2b1f",
	"1546961342-ea5f60ebe3a7", "1573496359142-b8d87734a5a2", "1499952127939-9bbf5af6c51c",
	"1569779213684-ff3b91e61200", "1573497019940-1c28c88b4f3e", "1590650046871-92c887180603",
	"1580489944761-15a19d674956", "1544005313-94ddf0286df2",
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

var danceStyles = []string{
	"salsa", "bachata", "kizomba", "tango", "waltz",
	"hip-hop", "contemporary", "latin", "ballroom", "swing",
	"zouk", "cha-cha", "rumba", "samba", "paso-doble",
}

var categories = []string{
	"salsa", "bachata", "kizomba", "tango", "waltz",
	"hip-hop", "contemporary", "latin",
}

var goals = []string{"hobby", "professional"}
var programs = []string{"standard", "latina", "both"}

var biosMale = []string{
	"Танцюю вже 5 років. Шукаю партнерку для тренувань і виступів.",
	"Люблю сальсу і бачату. Відкритий до нових знайомств у світі танцю.",
	"Танець — моя пристрасть. Хочу знайти однодумців.",
	"Серйозно займаюсь стандартними програмами. Шукаю партнера для пар.",
	"Танцюю для задоволення. Люблю соціальні танці.",
	"Починаючий танцюрист із великим бажанням розвиватись.",
	"Чемпіон міста з бальних танців. Шукаю партнерку для нового сезону.",
	"Танець — спосіб виразити себе. Відкритий до всіх стилів.",
}

var biosFemale = []string{
	"Танцюю класику і латину. Шукаю партнера для пар чи дуетних виступів.",
	"Захоплююсь бачатою та кізомбою. Рада познайомитись.",
	"Танець — це життя. Шукаю людей із такою ж пристрастю.",
	"Займаюсь хіп-хопом і контемпом. Відкрита до колаборацій.",
	"Люблю соціальні танці. Шукаю партнера для практик.",
	"Танцюю 8 років. Маю досвід у конкурсній програмі.",
	"Сальса — моя любов. Шукаю партнера для виступів.",
	"Захоплююсь танго і вальсом. Рада новим знайомствам.",
}

type cityEntry struct {
	city string
	lat  float64
	lng  float64
}

var ukraineCities = []cityEntry{
	{"Київ", 50.4501, 30.5234},
	{"Харків", 49.9935, 36.2304},
	{"Львів", 49.8397, 24.0297},
	{"Одеса", 46.4825, 30.7233},
	{"Дніпро", 48.4647, 35.0462},
	{"Запоріжжя", 47.8388, 35.1396},
	{"Вінниця", 49.2331, 28.4682},
	{"Полтава", 49.5883, 34.5514},
	{"Чернігів", 51.4982, 31.2893},
	{"Івано-Франківськ", 48.9226, 24.7111},
}

type profileData struct {
	firstName  string
	lastName   string
	gender     string
	birthDate  time.Time
	heightCm   int
	goal       string
	program    string
	danceStyles []string
	categories []string
	bio        string
	country    string
	city       string
	lat        float64
	lng        float64
}

func (p profileData) age() int {
	return int(time.Since(p.birthDate).Hours() / 8766)
}

func randomProfile(seed int, genderSplit string) profileData {
	r := rand.New(rand.NewSource(int64(seed) * 1337))

	gender := "male"
	switch genderSplit {
	case "female":
		gender = "female"
	case "male":
		gender = "male"
	default: // balanced
		if r.Intn(2) == 0 {
			gender = "female"
		}
	}

	firstName := firstNamesMale[r.Intn(len(firstNamesMale))]
	if gender == "female" {
		firstName = firstNamesFemale[r.Intn(len(firstNamesFemale))]
	}

	// Age 19-45
	age := 19 + r.Intn(27)
	birthYear := time.Now().Year() - age
	birthDate := time.Date(birthYear, time.Month(1+r.Intn(12)), 1+r.Intn(28), 0, 0, 0, 0, time.UTC)

	// 1-3 dance styles
	nds := 1 + r.Intn(3)
	ds := pickUnique(r, danceStyles, nds)

	// 1-3 categories
	nc := 1 + r.Intn(3)
	cats := pickUnique(r, categories, nc)

	city := ukraineCities[r.Intn(len(ukraineCities))]
	jitter := func() float64 { return (r.Float64() - 0.5) * 0.08 }

	var bio string
	if gender == "female" {
		bio = biosFemale[r.Intn(len(biosFemale))]
	} else {
		bio = biosMale[r.Intn(len(biosMale))]
	}

	return profileData{
		firstName:   firstName,
		lastName:    lastNames[r.Intn(len(lastNames))],
		gender:      gender,
		birthDate:   birthDate,
		heightCm:    150 + r.Intn(40), // 150-189 cm
		goal:        goals[r.Intn(len(goals))],
		program:     programs[r.Intn(len(programs))],
		danceStyles: ds,
		categories:  cats,
		bio:         bio,
		country:     "Україна",
		city:        city.city,
		lat:         city.lat + jitter(),
		lng:         city.lng + jitter(),
	}
}

func pickUnique(r *rand.Rand, pool []string, n int) []string {
	if n > len(pool) {
		n = len(pool)
	}
	used := map[int]bool{}
	out := make([]string, n)
	for i := 0; i < n; i++ {
		for {
			idx := r.Intn(len(pool))
			if !used[idx] {
				used[idx] = true
				out[i] = pool[idx]
				break
			}
		}
	}
	return out
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
