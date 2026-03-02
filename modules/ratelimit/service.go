package ratelimit

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	ratelimit "github.com/JGLTechnologies/gin-rate-limit"
	"github.com/redis/go-redis/v9"

	"github.com/Gooowan/matchup/modules/core/auth"
	"github.com/Gooowan/matchup/modules/core/utils"
	"github.com/gin-gonic/gin"
)

type RLService struct {
	LoginRateLimiter        gin.HandlerFunc
	ExchangeRateLimiter     gin.HandlerFunc
	WithdrawCodeRateLimiter gin.HandlerFunc
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

	exchangeStore := ratelimit.RedisStore(&ratelimit.RedisOptions{
		RedisClient: redisClient,
		Rate:        time.Minute,
		Limit:       10,
	})
	exchangeRateLimiter := ratelimit.RateLimiter(exchangeStore, &ratelimit.Options{
		ErrorHandler: errorHandler,
		KeyFunc:      exchangeRateKeyFunc,
	})

	withdrawCodeStore := ratelimit.RedisStore(&ratelimit.RedisOptions{
		RedisClient: redisClient,
		Rate:        time.Second,
		Limit:       10,
	})
	withdrawCodeRateLimiter := ratelimit.RateLimiter(withdrawCodeStore, &ratelimit.Options{
		ErrorHandler: errorHandler,
		KeyFunc:      userKeyFunc,
	})

	return &RLService{
		LoginRateLimiter:        loginRateLimiter,
		ExchangeRateLimiter:     exchangeRateLimiter,
		WithdrawCodeRateLimiter: withdrawCodeRateLimiter,
	}
}

func exchangeRateKeyFunc(c *gin.Context) string {
	user, ok := auth.GetUserFromContext(c)
	if !ok {
		return fmt.Sprintf("rl:exchange:%s", c.ClientIP())
	}
	return fmt.Sprintf("rl:exchange:%s", utils.UUIDToString(user.ID))
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
