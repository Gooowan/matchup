// cmd/seed-all creates a realistic multi-thousand-user dataset for load and UX
// testing. It inserts users, profiles, preferences, swipes, mutual-match chats,
// and starter messages directly into the database for maximum speed.
//
// Usage:
//
//	DATABASE_URL=postgres://... go run ./cmd/seed-all [flags]
//
// Flags:
//
//	--count         dancer profiles to create     (default 10000)
//	--swipes-each   swipes each user generates    (default 40)
//	--like-rate     fraction of swipes that LIKE  (default 0.62)
//	--msg-min       min messages per chat         (default 3)
//	--msg-max       max messages per chat         (default 12)
//	--password      plain-text password           (default "password123")
//	--batch         DB insert batch size          (default 500)
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// ── data pools ───────────────────────────────────────────────────────────────

var firstNamesMale = []string{
	"Олексій", "Іван", "Андрій", "Дмитро", "Микола",
	"Богдан", "Євген", "Роман", "Тарас", "Олег",
	"Михайло", "Павло", "Денис", "Юрій", "Віктор",
	"Сергій", "Василь", "Максим", "Артем", "Владислав",
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
	"Семенченко", "Захаренко", "Власенко", "Литвиненко", "Гриценко",
}
var goals    = []string{"hobby", "professional"}
var programs = []string{"standard", "latina", "both"}
var categories = []string{
	"kids", "juvenile1", "juvenile2", "junior1", "junior2", "youth", "adult",
}
var biosMale = []string{
	"Танцюю вже 5 років. Шукаю партнерку для тренувань і виступів.",
	"Люблю сальсу і бачату. Відкритий до нових знайомств у світі танцю.",
	"Танець — моя пристрасть. Хочу знайти однодумців.",
	"Серйозно займаюсь стандартними програмами. Шукаю партнера для пар.",
	"Танцюю для задоволення. Люблю соціальні танці.",
	"Починаючий танцюрист із великим бажанням розвиватись.",
	"Чемпіон міста з бальних танців. Шукаю партнерку для нового сезону.",
	"Танець — спосіб виразити себе. Відкритий до всіх стилів.",
	"Займаюсь танцями 3 роки. Мрію виступати на міжнародних турнірах.",
	"Шукаю серйозного партнера для підготовки до змагань.",
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
	"Тренуюсь щодня. Потрібен серйозний партнер для сезону.",
	"Люблю стандарт і латину. Відкрита до пропозицій.",
}

// Chat messages pool — realistic short dance-world messages.
var chatOpeners = []string{
	"Привіт! Бачив(ла) твій профіль — дуже круто танцюєш 🔥",
	"Вітаю! Шукаю партнера для тренувань, може поговоримо?",
	"Привіт! Цікаво познайомитись ближче 😊",
	"Привіт! Ти з якого клубу?",
	"Вітаю! Давно займаєшся танцями?",
	"Привіт! Бачила, що ти теж любиш стандарт — може потренуємось разом?",
	"Привіт! Яка у тебе категорія зараз?",
}
var chatReplies = []string{
	"Дякую! Танцюю вже 4 роки 😊",
	"Привіт! Так, звичайно, давай поговоримо.",
	"Привіт! Я з клубу «Реверанс», а ти?",
	"Дуже приємно! Теж шукаю партнера.",
	"Привіт! Зараз юніор-1, працюємо над стандартом.",
	"Звучить цікаво! Коли зазвичай тренуєшся?",
	"Дякую за лайк! Може спробуємо потренуватись разом?",
	"Привіт! Так, займаюсь уже 6 років.",
}
var chatMessages = []string{
	"Коли зручно зустрітись на тренування?",
	"Ти на які змагання готуєшся цього сезону?",
	"Мій тренер каже, що нам треба більше практикувати лайндансинг.",
	"Як часто тренуєшся на тиждень?",
	"Був(ла) на минулому турнірі? Як пройшло?",
	"Може завтра потренуємось після 18:00?",
	"У мене є вільний час у вівторок і четвер — тебе влаштовує?",
	"Чудово! Жду твого повідомлення щодо розкладу.",
	"Також шукаю хорошого партнера вже кілька місяців.",
	"Давай спробуємо. Котра зручна для першого тренування?",
	"Відмінно! Напиши адресу залу.",
	"Добре. Домовились на п'ятницю о 19:00?",
	"Супер! До зустрічі 🙌",
	"Дякую, теж рада познайомитись!",
	"Чудово виходить, продовжуємо? 😄",
}

type cityEntry struct {
	name string
	lat  float64
	lng  float64
}

var kyivCities = []cityEntry{
	{"Київ", 50.4501, 30.5234},
}

// ── main ─────────────────────────────────────────────────────────────────────

func main() {
	count     := flag.Int("count", 10000, "number of dancer profiles to create")
	swipesEach := flag.Int("swipes-each", 40, "swipes each user generates")
	likeRate  := flag.Float64("like-rate", 0.62, "fraction of swipes that are LIKE")
	msgMin    := flag.Int("msg-min", 3, "min messages per matched chat")
	msgMax    := flag.Int("msg-max", 12, "max messages per matched chat")
	password  := flag.String("password", "password123", "plain-text password for all seeded accounts")
	batchSize := flag.Int("batch", 500, "insert batch size")
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
		fatalf("db connect: %v", err)
	}
	defer pool.Close()

	// Pre-compute bcrypt hash once (expensive; do not repeat per user).
	fmt.Printf("⏳ Hashing password…\n")
	hashBytes, err := bcrypt.GenerateFromPassword([]byte(*password), bcrypt.DefaultCost)
	if err != nil {
		fatalf("bcrypt: %v", err)
	}
	hash := string(hashBytes)

	// Load active Kyiv clubs.
	type clubRow struct{ id string; lat, lng float64 }
	var clubs []clubRow
	rows, err := pool.Query(ctx, `
		SELECT id::text, latitude, longitude
		FROM clubs
		WHERE is_active = true AND city IN ('Київ','Kyiv')
		  AND latitude != 0 AND longitude != 0`)
	if err != nil {
		fatalf("load clubs: %v", err)
	}
	for rows.Next() {
		var c clubRow
		if e := rows.Scan(&c.id, &c.lat, &c.lng); e == nil {
			clubs = append(clubs, c)
		}
	}
	rows.Close()
	if len(clubs) == 0 {
		fatalf("no active Kyiv clubs found — create at least one via the app first")
	}
	fmt.Printf("ℹ  %d Kyiv clubs available\n", len(clubs))

	// ── Phase 1: Users + profiles + preferences ───────────────────────────────
	fmt.Printf("\n=== Phase 1/4: users + profiles + preferences (target: %d) ===\n", *count)

	// Collect newly inserted IDs for swipe/chat generation.
	var userIDs []string
	created, skipped := 0, 0

	for batch := 0; batch*(*batchSize) < *count; batch++ {
		start := batch * (*batchSize)
		end   := min(start+*batchSize, *count)

		type row struct {
			idx  int
			email string
		}
		var toInsert []row
		for i := start + 1; i <= end; i++ {
			email := fmt.Sprintf("seed-%05d@matchup.local", i)
			var exists bool
			if e := pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE email=$1)`, email).Scan(&exists); e != nil {
				fatalf("check exists %d: %v", i, e)
			}
			if exists {
				// Collect existing ID so we can still generate swipes for it.
				var id string
				if e2 := pool.QueryRow(ctx, `SELECT id::text FROM users WHERE email=$1`, email).Scan(&id); e2 == nil {
					userIDs = append(userIDs, id)
				}
				skipped++
				continue
			}
			toInsert = append(toInsert, row{i, email})
		}

		for _, r := range toInsert {
			p := randomProfile(r.idx)
			club := clubs[rand.Intn(len(clubs))]

			pdJSON, _ := json.Marshal(map[string]any{
				"first_name": p.firstName,
				"last_name":  p.lastName,
				"avatar":     placeholderAvatar(r.idx, p.gender),
			})
			metaJSON, _ := json.Marshal(map[string]any{
				"bio":        p.bio,
				"media_urls": placeholderGallery(r.idx),
				"account_type": "dancer",
			})

			var userID string
			if e := pool.QueryRow(ctx, `
				INSERT INTO users (email, password, role, profile_data)
				VALUES ($1,$2,'USER',$3::jsonb)
				RETURNING id::text
			`, r.email, hash, string(pdJSON)).Scan(&userID); e != nil {
				fatalf("insert user %d: %v", r.idx, e)
			}

			oppGender := "female"
			if p.gender == "female" { oppGender = "male" }

			cats := pickN(rand.New(rand.NewSource(int64(r.idx))), categories, 1+rand.Intn(3))

			_, e2 := pool.Exec(ctx, `
				INSERT INTO profiles
					(user_id, account_type, gender, birth_date, height_cm,
					 goal, program, categories, country, city,
					 latitude, longitude, visible, primary_club_id, metadata)
				VALUES ($1,'dancer',$2,$3::date,$4,$5,$6,$7::text[],
					'Україна','Київ',$8,$9,true,$10::uuid,$11::jsonb)
			`, userID, p.gender,
				p.birthDate.Format("2006-01-02"), p.heightCm,
				p.goal, p.program, pgArr(cats),
				club.lat+jitter(), club.lng+jitter(),
				club.id, string(metaJSON),
			)
			if e2 != nil { fatalf("insert profile %d: %v", r.idx, e2) }

			_, e3 := pool.Exec(ctx, `
				INSERT INTO user_preferences
					(user_id, preferred_gender, age_min, age_max,
					 preferred_country, preferred_city)
				VALUES ($1,$2,$3,$4,'Україна','Київ')
			`, userID, oppGender, p.age()-8, p.age()+8)
			if e3 != nil { fatalf("insert prefs %d: %v", r.idx, e3) }

			_, _ = pool.Exec(ctx, `
				INSERT INTO club_members (club_id, user_id, role)
				VALUES ($1::uuid,$2::uuid,'member')
				ON CONFLICT DO NOTHING
			`, club.id, userID)

			userIDs = append(userIDs, userID)
			created++
		}

		done := min((batch+1)*(*batchSize), *count)
		fmt.Printf("  profiles %d/%d (created %d, skipped %d)\n", done, *count, created, skipped)
	}
	fmt.Printf("✓ Phase 1 done. %d created, %d skipped.\n", created, skipped)

	n := len(userIDs)
	if n < 2 {
		fmt.Println("Fewer than 2 users — skipping swipe/chat phases.")
		return
	}

	// ── Phase 2: Swipes (LIKE / PASS) ────────────────────────────────────────
	fmt.Printf("\n=== Phase 2/4: swipes (%d users × %d swipes) ===\n", n, *swipesEach)

	// liked[a][b] = true means a LIKED b
	liked := make(map[string]map[string]bool, n)
	for _, id := range userIDs {
		liked[id] = make(map[string]bool)
	}

	swipeTotal, likeTotal := 0, 0
	swipeBuf := make([][]any, 0, *batchSize)

	flushSwipes := func() {
		if len(swipeBuf) == 0 { return }
		// Build batch insert
		placeholders := make([]string, len(swipeBuf))
		args := make([]any, 0, len(swipeBuf)*4)
		for i, row := range swipeBuf {
			base := i * 4
			placeholders[i] = fmt.Sprintf("($%d,$%d,$%d,$%d)", base+1, base+2, base+3, base+4)
			args = append(args, row...)
		}
		_, err := pool.Exec(ctx, `
			INSERT INTO matches (from_user_id, to_user_id, action, source)
			VALUES `+strings.Join(placeholders, ",")+`
			ON CONFLICT (from_user_id, to_user_id) DO NOTHING
		`, args...)
		if err != nil { fatalf("insert swipes batch: %v", err) }
		swipeBuf = swipeBuf[:0]
	}

	r := rand.New(rand.NewSource(42))
	for i, fromID := range userIDs {
		// Pick `swipesEach` unique targets (skip self).
		targets := pickIDsExcluding(r, userIDs, *swipesEach, i)
		for _, toID := range targets {
			action := "PASS"
			if r.Float64() < *likeRate {
				action = "LIKE"
				liked[fromID][toID] = true
				likeTotal++
			}
			swipeBuf = append(swipeBuf, []any{fromID, toID, action, "feed"})
			swipeTotal++
			if len(swipeBuf) >= *batchSize {
				flushSwipes()
			}
		}
		if (i+1)%1000 == 0 {
			flushSwipes()
			fmt.Printf("  swipes: %d/%d users processed\n", i+1, n)
		}
	}
	flushSwipes()
	fmt.Printf("✓ Phase 2 done. %d swipes inserted (%d likes).\n", swipeTotal, likeTotal)

	// ── Phase 3: Mutual matches → chats ──────────────────────────────────────
	fmt.Printf("\n=== Phase 3/4: detecting mutual matches → chats ===\n")

	type chatPair struct{ u1, u2 string }
	var pairs []chatPair
	for _, a := range userIDs {
		for b := range liked[a] {
			if liked[b][a] && a < b { // deduplicate
				pairs = append(pairs, chatPair{a, b})
			}
		}
	}
	fmt.Printf("  %d mutual matches found → creating chats…\n", len(pairs))

	chatIDs := make([]string, 0, len(pairs))
	chatBuf := make([][]any, 0, *batchSize)

	flushChats := func() {
		if len(chatBuf) == 0 { return }
		ph := make([]string, len(chatBuf))
		args := make([]any, 0, len(chatBuf)*2)
		for i, row := range chatBuf {
			ph[i] = fmt.Sprintf("($%d,$%d)", i*2+1, i*2+2)
			args = append(args, row...)
		}
		rows2, err := pool.Query(ctx, `
			INSERT INTO chats (user1_id, user2_id)
			VALUES `+strings.Join(ph, ",")+`
			ON CONFLICT DO NOTHING
			RETURNING id::text
		`, args...)
		if err != nil { fatalf("insert chats: %v", err) }
		for rows2.Next() {
			var id string
			_ = rows2.Scan(&id)
			chatIDs = append(chatIDs, id)
		}
		rows2.Close()
		chatBuf = chatBuf[:0]
	}

	chatOwner := make(map[string]string) // chatID → user1_id (opener)
	for _, p := range pairs {
		chatOwner[p.u1+"|"+p.u2] = p.u1
		chatBuf = append(chatBuf, []any{p.u1, p.u2})
		if len(chatBuf) >= *batchSize {
			flushChats()
		}
	}
	flushChats()
	_ = chatOwner
	fmt.Printf("✓ Phase 3 done. %d chats created.\n", len(chatIDs))

	// ── Phase 4: Messages ─────────────────────────────────────────────────────
	fmt.Printf("\n=== Phase 4/4: seeding messages (%d chats) ===\n", len(chatIDs))

	// Re-query chat rows to get user1_id / user2_id for each chat.
	type chatInfo struct{ id, u1, u2 string }
	var chatInfos []chatInfo
	for i := 0; i < len(chatIDs); i += 200 {
		end := min(i+200, len(chatIDs))
		chunk := chatIDs[i:end]
		ph := make([]string, len(chunk))
		args := make([]any, len(chunk))
		for j, id := range chunk {
			ph[j] = fmt.Sprintf("$%d", j+1)
			args[j] = id
		}
		rs, err := pool.Query(ctx, `
			SELECT id::text, user1_id::text, user2_id::text
			FROM chats WHERE id = ANY(ARRAY[`+strings.Join(ph, ",")+`]::uuid[])
		`, args...)
		if err != nil { fatalf("query chats: %v", err) }
		for rs.Next() {
			var c chatInfo
			_ = rs.Scan(&c.id, &c.u1, &c.u2)
			chatInfos = append(chatInfos, c)
		}
		rs.Close()
	}

	msgBuf := make([][]any, 0, *batchSize)
	msgTotal := 0

	flushMsgs := func() {
		if len(msgBuf) == 0 { return }
		ph := make([]string, len(msgBuf))
		args := make([]any, 0, len(msgBuf)*4)
		for i, row := range msgBuf {
			base := i * 4
			ph[i] = fmt.Sprintf("($%d,$%d,$%d,$%d)", base+1, base+2, base+3, base+4)
			args = append(args, row...)
		}
		_, err := pool.Exec(ctx, `
			INSERT INTO messages (chat_id, sender_id, type, content)
			VALUES `+strings.Join(ph, ","), args...)
		if err != nil { fatalf("insert messages: %v", err) }
		msgBuf = msgBuf[:0]
	}

	for idx, ci := range chatInfos {
		msgCount := *msgMin + r.Intn(*msgMax-*msgMin+1)
		// Opener (user1 starts)
		msgBuf = append(msgBuf, []any{ci.id, ci.u1, "TEXT", chatOpeners[r.Intn(len(chatOpeners))]})
		msgTotal++
		// Reply from user2
		if msgCount > 1 {
			msgBuf = append(msgBuf, []any{ci.id, ci.u2, "TEXT", chatReplies[r.Intn(len(chatReplies))]})
			msgTotal++
		}
		// Subsequent messages alternating
		sender := ci.u1
		for m := 2; m < msgCount; m++ {
			if m%2 == 0 { sender = ci.u2 } else { sender = ci.u1 }
			msgBuf = append(msgBuf, []any{ci.id, sender, "TEXT", chatMessages[r.Intn(len(chatMessages))]})
			msgTotal++
		}
		if len(msgBuf) >= *batchSize {
			flushMsgs()
		}
		if (idx+1)%5000 == 0 {
			fmt.Printf("  messages: %d/%d chats done\n", idx+1, len(chatInfos))
		}
	}
	flushMsgs()
	fmt.Printf("✓ Phase 4 done. %d messages inserted.\n", msgTotal)

	// ── Summary ───────────────────────────────────────────────────────────────
	fmt.Printf(`
╔══════════════════════════════════════╗
║          seed-all complete           ║
╠══════════════════════════════════════╣
║  Users created   : %6d            ║
║  Users skipped   : %6d            ║
║  Swipes          : %6d            ║
║  Mutual matches  : %6d            ║
║  Chats           : %6d            ║
║  Messages        : %6d            ║
╚══════════════════════════════════════╝
`, created, skipped, swipeTotal, len(pairs), len(chatIDs), msgTotal)
}

// ── helpers ───────────────────────────────────────────────────────────────────

type profile struct {
	firstName string
	lastName  string
	gender    string
	birthDate time.Time
	heightCm  int
	goal      string
	program   string
	bio       string
}

func (p profile) age() int {
	return int(time.Since(p.birthDate).Hours() / 8766)
}

func randomProfile(seed int) profile {
	r := rand.New(rand.NewSource(int64(seed) * 1337))

	gender := "male"
	if r.Intn(2) == 0 { gender = "female" }

	firstName := firstNamesMale[r.Intn(len(firstNamesMale))]
	if gender == "female" { firstName = firstNamesFemale[r.Intn(len(firstNamesFemale))] }

	age := 19 + r.Intn(26)
	birthDate := time.Date(
		time.Now().Year()-age,
		time.Month(1+r.Intn(12)),
		1+r.Intn(28), 0, 0, 0, 0, time.UTC,
	)

	bio := biosMale[r.Intn(len(biosMale))]
	if gender == "female" { bio = biosFemale[r.Intn(len(biosFemale))] }

	return profile{
		firstName: firstName,
		lastName:  lastNames[r.Intn(len(lastNames))],
		gender:    gender,
		birthDate: birthDate,
		heightCm:  155 + r.Intn(35),
		goal:      goals[r.Intn(len(goals))],
		program:   programs[r.Intn(len(programs))],
		bio:       bio,
	}
}

func placeholderAvatar(seed int, gender string) string {
	// Deterministic Unsplash face photos (no network call at seed time).
	maleSeeds   := []int{1526413232644, 1519238951851, 1500648174389, 1507003107896, 1492702298905}
	femaleSeeds := []int{1494790408171, 1506746990176, 1529727888906, 1540569484055, 1520005507060}
	pool := maleSeeds
	if gender == "female" { pool = femaleSeeds }
	s := pool[seed%len(pool)]
	return fmt.Sprintf("https://images.unsplash.com/photo-%d?w=400&q=80&auto=format&fit=crop", s)
}

func placeholderGallery(seed int) []string {
	bases := []int{1526413232644, 1500648174389, 1519238951851}
	out := make([]string, 2)
	for i := range out {
		out[i] = fmt.Sprintf("https://images.unsplash.com/photo-%d?w=600&q=80&auto=format&fit=crop&sig=%d",
			bases[(seed+i)%len(bases)], seed+i)
	}
	return out
}

func pickIDsExcluding(r *rand.Rand, ids []string, n, excludeIdx int) []string {
	if n >= len(ids)-1 { n = len(ids) - 1 }
	picked := make(map[int]bool)
	out := make([]string, 0, n)
	for len(out) < n {
		idx := r.Intn(len(ids))
		if idx == excludeIdx || picked[idx] { continue }
		picked[idx] = true
		out = append(out, ids[idx])
	}
	return out
}

func pickN(r *rand.Rand, pool []string, n int) []string {
	if n > len(pool) { n = len(pool) }
	used := map[int]bool{}
	out := make([]string, n)
	for i := 0; i < n; i++ {
		for { idx := r.Intn(len(pool)); if !used[idx] { used[idx]=true; out[i]=pool[idx]; break } }
	}
	return out
}

func pgArr(ss []string) string {
	if len(ss) == 0 { return "{}" }
	quoted := make([]string, len(ss))
	for i, s := range ss { quoted[i] = `"` + s + `"` }
	return "{" + strings.Join(quoted, ",") + "}"
}

func jitter() float64 { return (rand.Float64() - 0.5) * 0.06 }

func min(a, b int) int { if a < b { return a }; return b }

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" { return v }
	return def
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "seed-all: "+format+"\n", args...)
	os.Exit(1)
}
