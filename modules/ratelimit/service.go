package ratelimit

import (
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

type RLService struct {
	LoginRateLimiter   gin.HandlerFunc
	SwipeRateLimiter   gin.HandlerFunc
	MessageRateLimiter gin.HandlerFunc
	UploadRateLimiter  gin.HandlerFunc
	RegisterRateLimiter gin.HandlerFunc
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
