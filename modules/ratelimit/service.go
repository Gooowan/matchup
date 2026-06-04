package ratelimit

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	ratelimit "github.com/JGLTechnologies/gin-rate-limit"
	"github.com/redis/go-redis/v9"

	"github.com/Gooowan/matchup/modules/core/utils"
	"github.com/Gooowan/matchup/modules/users/auth"
	"github.com/gin-gonic/gin"
)

const (
	loginLockoutMaxAttempts = 10
	loginLockoutDuration    = 15 * time.Minute
)

type RLService struct {
	LoginRateLimiter    gin.HandlerFunc
	SwipeRateLimiter    gin.HandlerFunc
	MessageRateLimiter  gin.HandlerFunc
	UploadRateLimiter   gin.HandlerFunc
	RegisterRateLimiter gin.HandlerFunc
	// redis client held for lockout helpers
	redis *redis.Client
}

// RecordLoginFailure increments the failure counter for an email address and
// sets a 15-minute lockout once the threshold is reached. Call this after a
// failed authentication attempt.
func (s *RLService) RecordLoginFailure(ctx context.Context, email string) {
	key := fmt.Sprintf("lockout:login:%s", strings.ToLower(strings.TrimSpace(email)))
	count, _ := s.redis.Incr(ctx, key).Result()
	if count == 1 {
		// First failure — start the sliding window.
		s.redis.Expire(ctx, key, loginLockoutDuration)
	}
	if count >= loginLockoutMaxAttempts {
		// Re-set TTL to lockout duration each time we reach/exceed the threshold.
		s.redis.Expire(ctx, key, loginLockoutDuration)
	}
}

// ClearLoginFailures removes any lockout counter for an email after a
// successful login, preventing phantom lockouts.
func (s *RLService) ClearLoginFailures(ctx context.Context, email string) {
	key := fmt.Sprintf("lockout:login:%s", strings.ToLower(strings.TrimSpace(email)))
	s.redis.Del(ctx, key)
}

// LoginLockoutMiddleware rejects requests for accounts that have exceeded
// loginLockoutMaxAttempts failed logins within the lockout window. This runs
// before the per-minute sliding-window rate limiter so locked accounts are
// rejected immediately without burning the rate-limit budget.
func (s *RLService) LoginLockoutMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		body, err := c.GetRawData()
		if err != nil {
			c.Next()
			return
		}
		c.Request.Body = io.NopCloser(strings.NewReader(string(body)))

		var req struct {
			Email string `json:"email"`
		}
		_ = json.Unmarshal(body, &req)
		if req.Email == "" {
			c.Next()
			return
		}

		key := fmt.Sprintf("lockout:login:%s", strings.ToLower(strings.TrimSpace(req.Email)))
		count, _ := s.redis.Get(c.Request.Context(), key).Int64()
		if count >= loginLockoutMaxAttempts {
			ttl, _ := s.redis.TTL(c.Request.Context(), key).Result()
			c.Header("Retry-After", fmt.Sprintf("%.0f", ttl.Seconds()))
			c.JSON(429, gin.H{
				"error":      "Account temporarily locked due to too many failed login attempts. Try again later.",
				"error_code": "ACCOUNT_LOCKED",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

func NewRLService(redisClient *redis.Client) *RLService {
	loginStore := ratelimit.RedisStore(&ratelimit.RedisOptions{
		RedisClient: redisClient,
		Rate:        time.Minute,
		Limit:       3,
	})
	loginRateLimiter := ratelimit.RateLimiter(loginStore, &ratelimit.Options{
		ErrorHandler: errorHandler,
		KeyFunc:      emailKeyFunc,
	})

	swipeStore := ratelimit.RedisStore(&ratelimit.RedisOptions{
		RedisClient: redisClient,
		Rate:        time.Minute,
		Limit:       200,
	})
	swipeRateLimiter := ratelimit.RateLimiter(swipeStore, &ratelimit.Options{
		ErrorHandler: errorHandler,
		KeyFunc:      userKeyFunc,
	})

	messageStore := ratelimit.RedisStore(&ratelimit.RedisOptions{
		RedisClient: redisClient,
		Rate:        time.Minute,
		Limit:       60,
	})
	messageRateLimiter := ratelimit.RateLimiter(messageStore, &ratelimit.Options{
		ErrorHandler: errorHandler,
		KeyFunc:      userKeyFunc,
	})

	uploadStore := ratelimit.RedisStore(&ratelimit.RedisOptions{
		RedisClient: redisClient,
		Rate:        24 * time.Hour,
		Limit:       20,
	})
	uploadRateLimiter := ratelimit.RateLimiter(uploadStore, &ratelimit.Options{
		ErrorHandler: errorHandler,
		KeyFunc:      userKeyFunc,
	})

	registerStore := ratelimit.RedisStore(&ratelimit.RedisOptions{
		RedisClient: redisClient,
		Rate:        time.Hour,
		Limit:       5,
	})
	registerRateLimiter := ratelimit.RateLimiter(registerStore, &ratelimit.Options{
		ErrorHandler: errorHandler,
		KeyFunc: func(c *gin.Context) string {
			return fmt.Sprintf("rl:register:%s", c.ClientIP())
		},
	})

	return &RLService{
		LoginRateLimiter:    loginRateLimiter,
		SwipeRateLimiter:    swipeRateLimiter,
		MessageRateLimiter:  messageRateLimiter,
		UploadRateLimiter:   uploadRateLimiter,
		RegisterRateLimiter: registerRateLimiter,
		redis:               redisClient,
	}
}

func userKeyFunc(c *gin.Context) string {
	user, ok := auth.GetUserFromContext(c)
	if !ok {
		return fmt.Sprintf("rl:user:%s", c.ClientIP())
	}
	return fmt.Sprintf("rl:user:%s", utils.UUIDToString(user.ID))
}

func emailKeyFunc(c *gin.Context) string {
	var req struct {
		Email string `json:"email"`
	}

	body, err := c.GetRawData()
	if err != nil {
		return fmt.Sprintf("rl:email:%s", c.ClientIP())
	}

	c.Request.Body = io.NopCloser(strings.NewReader(string(body)))

	if err := json.Unmarshal(body, &req); err != nil {
		return fmt.Sprintf("rl:email:%s", c.ClientIP())
	}

	if req.Email == "" {
		return fmt.Sprintf("rl:email:%s", c.ClientIP())
	}

	return fmt.Sprintf("rl:email:%s", strings.ToLower(strings.TrimSpace(req.Email)))
}

func errorHandler(c *gin.Context, info ratelimit.Info) {
	remainingTime := time.Until(info.ResetTime)
	c.Header("Retry-After", fmt.Sprintf("%.0f", remainingTime.Seconds()))
	c.JSON(429, gin.H{
		"error":               "Too many requests. Please try again later.",
		"retry_after":         remainingTime.String(),
		"retry_after_seconds": int(remainingTime.Seconds()),
	})
}
