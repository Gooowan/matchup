package push

import (
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/payload"
	"github.com/sideshow/apns2/token"
)

type Service struct {
	db     *pgxpool.Pool
	client *apns2.Client
	bundleID string
	mu     sync.Mutex
	log    *slog.Logger
}

func NewService(db *pgxpool.Pool, log *slog.Logger) *Service {
	svc := &Service{
		db:       db,
		bundleID: os.Getenv("APNS_BUNDLE_ID"),
		log:      log,
	}

	keyPath := os.Getenv("APNS_KEY_PATH")
	keyID := os.Getenv("APNS_KEY_ID")
	teamID := os.Getenv("APNS_TEAM_ID")

	if keyPath != "" && keyID != "" && teamID != "" {
		if authKey, err := loadKey(keyPath); err == nil {
			tok := &token.Token{
				AuthKey: authKey,
				KeyID:   keyID,
				TeamID:  teamID,
			}
			if os.Getenv("APNS_ENV") == "production" {
				svc.client = apns2.NewTokenClient(tok).Production()
			} else {
				svc.client = apns2.NewTokenClient(tok).Development()
			}
			log.Info("APNs client initialized", "env", os.Getenv("APNS_ENV"))
		} else {
			log.Warn("APNs key load failed — push notifications disabled", "error", err)
		}
	} else {
		log.Info("APNs env vars not set — push notifications disabled")
	}

	return svc
}

func loadKey(path string) (*ecdsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("no PEM block found in key file")
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	ec, ok := key.(*ecdsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("key is not ECDSA")
	}
	return ec, nil
}

// RegisterToken saves a device push token for a user.
func (s *Service) RegisterToken(ctx context.Context, userID, pushToken, platform string) error {
	_, err := s.db.Exec(ctx,
		`INSERT INTO user_push_tokens (user_id, token, platform)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (user_id, token) DO NOTHING`,
		userID, pushToken, platform,
	)
	return err
}

// SendToUser sends a push notification to all registered tokens for a user.
func (s *Service) SendToUser(ctx context.Context, userID string, title, body string) {
	if s.client == nil {
		s.log.Debug("push skipped (client not configured)", "user_id", userID, "title", title)
		return
	}

	rows, err := s.db.Query(ctx,
		`SELECT token FROM user_push_tokens WHERE user_id = $1 AND platform = 'ios'`,
		userID,
	)
	if err != nil {
		s.log.Error("push: failed to fetch tokens", "user_id", userID, "error", err)
		return
	}
	defer rows.Close()

	p := payload.NewPayload().AlertTitle(title).AlertBody(body).Sound("default")

	for rows.Next() {
		var deviceToken string
		if err := rows.Scan(&deviceToken); err != nil {
			continue
		}
		notification := &apns2.Notification{
			DeviceToken: deviceToken,
			Topic:       s.bundleID,
			Payload:     p,
		}
		res, err := s.client.PushWithContext(ctx, notification)
		if err != nil {
			s.log.Error("apns push failed", "user_id", userID, "error", err)
		} else if res.StatusCode != 200 {
			s.log.Warn("apns push rejected", "user_id", userID, "reason", res.Reason)
		}
	}
}
