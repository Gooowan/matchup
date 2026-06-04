// cmd/delete-user fully removes a user from the database in a single transaction.
// It deletes all dependent rows in the correct order before removing the user record.
// Clubs owned by the user are NOT deleted (owner_user_id is SET NULL by the schema);
// the tool prints a warning listing them so an admin can reassign ownership.
//
// Usage:
//
//	DATABASE_URL=postgres://... go run ./cmd/delete-user --email=foo@example.com
//	DATABASE_URL=postgres://... go run ./cmd/delete-user --id=<uuid> --yes
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	email := flag.String("email", "", "email of the user to delete")
	id := flag.String("id", "", "UUID of the user to delete")
	yes := flag.Bool("yes", false, "skip confirmation prompt")
	flag.Parse()

	if *email == "" && *id == "" {
		fmt.Fprintln(os.Stderr, "delete-user: provide --email or --id")
		flag.Usage()
		os.Exit(1)
	}
	if *email != "" && *id != "" {
		fmt.Fprintln(os.Stderr, "delete-user: --email and --id are mutually exclusive")
		os.Exit(1)
	}

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

	// --- Resolve user ---
	var userID, userEmail, userRole string
	var query string
	var arg any
	if *email != "" {
		query = `SELECT id::text, email, role FROM users WHERE email = $1`
		arg = *email
	} else {
		query = `SELECT id::text, email, role FROM users WHERE id = $1::uuid`
		arg = *id
	}
	err = pool.QueryRow(ctx, query, arg).Scan(&userID, &userEmail, &userRole)
	if err == pgx.ErrNoRows {
		fatalf("user not found")
	}
	if err != nil {
		fatalf("lookup: %v", err)
	}

	fmt.Printf("User found:\n  ID:    %s\n  Email: %s\n  Role:  %s\n\n", userID, userEmail, userRole)

	// --- Warn about owned clubs ---
	rows, err := pool.Query(ctx, `SELECT id::text, name, slug FROM clubs WHERE owner_user_id = $1::uuid`, userID)
	if err != nil {
		fatalf("query clubs: %v", err)
	}
	var ownedClubs []string
	for rows.Next() {
		var cid, cname, cslug string
		_ = rows.Scan(&cid, &cname, &cslug)
		ownedClubs = append(ownedClubs, fmt.Sprintf("  %s  %-40s  /%s", cid, cname, cslug))
	}
	rows.Close()

	if len(ownedClubs) > 0 {
		fmt.Printf("⚠  This user owns %d club(s). Their owner_user_id will be set NULL:\n", len(ownedClubs))
		for _, c := range ownedClubs {
			fmt.Println(c)
		}
		fmt.Println()
	}

	// --- Media URLs (for manual S3 cleanup) ---
	mediaRows, err := pool.Query(ctx, `SELECT file_key FROM media WHERE owner_id = $1::uuid`, userID)
	if err != nil {
		fatalf("query media: %v", err)
	}
	var mediaKeys []string
	for mediaRows.Next() {
		var k string
		_ = mediaRows.Scan(&k)
		mediaKeys = append(mediaKeys, k)
	}
	mediaRows.Close()

	if len(mediaKeys) > 0 {
		fmt.Printf("ℹ  %d media object(s) referenced (NOT deleted from object storage):\n", len(mediaKeys))
		for _, k := range mediaKeys {
			fmt.Printf("  %s\n", k)
		}
		fmt.Println()
	}

	// --- Confirm ---
	if !*yes {
		fmt.Printf("Delete user %s (%s) and ALL their data? [y/N] ", userEmail, userID)
		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "y" && answer != "yes" {
			fmt.Println("Aborted.")
			os.Exit(0)
		}
	}

	// --- Delete in transaction ---
	tx, err := pool.Begin(ctx)
	if err != nil {
		fatalf("begin tx: %v", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	steps := []struct {
		desc string
		sql  string
	}{
		{"user_subscriptions",        `DELETE FROM user_subscriptions WHERE user_id = $1::uuid`},
		{"reports (reporter)",         `DELETE FROM reports WHERE reporter_id = $1::uuid`},
		{"reports (reported)",         `DELETE FROM reports WHERE reported_id = $1::uuid`},
		{"blocks",                     `DELETE FROM blocks WHERE blocker_id = $1::uuid OR blocked_id = $1::uuid`},
		{"messages (sender)",          `DELETE FROM messages WHERE sender_id = $1::uuid`},
		{"messages (chats)",           `DELETE FROM messages WHERE chat_id IN (SELECT id FROM chats WHERE user1_id = $1::uuid OR user2_id = $1::uuid)`},
		{"chats",                      `DELETE FROM chats WHERE user1_id = $1::uuid OR user2_id = $1::uuid`},
		{"matches",                    `DELETE FROM matches WHERE from_user_id = $1::uuid OR to_user_id = $1::uuid`},
		{"recommendation_likes_log",   `DELETE FROM recommendation_likes_log WHERE user_id = $1::uuid OR liked_id = $1::uuid`},
		{"user_locations",             `DELETE FROM user_locations WHERE user_id = $1::uuid`},
		{"media",                      `DELETE FROM media WHERE owner_id = $1::uuid`},
		{"user_preferences",           `DELETE FROM user_preferences WHERE user_id = $1::uuid`},
		{"profiles",                   `DELETE FROM profiles WHERE user_id = $1::uuid`},
		{"club_members",               `DELETE FROM club_members WHERE user_id = $1::uuid`},
		// Clear inviter_id FK — no ON DELETE clause, so must be nullified before
		// removing the user row or the FK constraint will reject the delete.
		{"users (clear inviter refs)", `UPDATE users SET inviter_id = NULL WHERE inviter_id = $1::uuid`},
		{"users",                      `DELETE FROM users WHERE id = $1::uuid`},
	}

	for _, step := range steps {
		tag, execErr := tx.Exec(ctx, step.sql, userID)
		if execErr != nil {
			err = execErr
			fatalf("delete %s: %v", step.desc, execErr)
		}
		if tag.RowsAffected() > 0 {
			fmt.Printf("  ✓ deleted %d row(s) from %s\n", tag.RowsAffected(), step.desc)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		fatalf("commit: %v", err)
	}

	fmt.Printf("\n✅ User %s (%s) has been fully deleted.\n", userEmail, userID)
	if len(ownedClubs) > 0 {
		fmt.Println("⚠  Remember to reassign ownership of the orphaned clubs listed above.")
	}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "delete-user: "+format+"\n", args...)
	os.Exit(1)
}
