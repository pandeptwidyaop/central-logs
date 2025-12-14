package middleware

import (
	"context"
	"strconv"
	"time"

	"central-logs/internal/queue"

	"github.com/gofiber/fiber/v2"
)

type RateLimitMiddleware struct {
	limiter *queue.RateLimiter
	limit   int
}

func NewRateLimitMiddleware(limiter *queue.RateLimiter, limit int) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		limiter: limiter,
		limit:   limit,
	}
}

// RateLimitByProject limits requests per project
func (m *RateLimitMiddleware) RateLimitByProject() fiber.Handler {
	return func(c *fiber.Ctx) error {
		project := GetProject(c)
		if project == nil {
			return c.Next()
		}

		ctx := context.Background()
		allowed, remaining, resetTime, err := m.limiter.AllowAPI(ctx, project.ID, m.limit)
		if err != nil {
			// Log error but don't block request
			return c.Next()
		}

		// Set rate limit headers
		c.Set("X-RateLimit-Limit", strconv.Itoa(m.limit))
		c.Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Set("X-RateLimit-Reset", strconv.FormatInt(resetTime.Unix(), 10))

		if !allowed {
			c.Set("Retry-After", strconv.FormatInt(int64(time.Until(resetTime).Seconds()), 10))
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":       "Rate limit exceeded",
				"retry_after": int(time.Until(resetTime).Seconds()),
			})
		}

		return c.Next()
	}
}
