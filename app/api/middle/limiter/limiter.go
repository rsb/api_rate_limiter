// Package limiter uses the foundation limits package to implement
// rate limiting middleware
package limiter

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rsb/api_rate_limiter/foundation/limits"
	"github.com/rsb/failure"
	"strconv"
	"time"
)

const (
	HeaderRateLimitLimit     = "X-RateLimit-Limit"
	HeaderRateLimitRemaining = "X-RateLimit-Remaining"
	HeaderRateLimitReset     = "X-RateLimit-Reset"
	HeaderRetryAfter         = "Retry-After"
)

func New(opts ...Config) fiber.Handler {
	cfg := configure(opts...)

	store := limits.NewMemoryStore(ToLimitsConfig(cfg))
	go store.GarbageCollector()
	return func(c *fiber.Ctx) error {
		if cfg.Next != nil && cfg.Next(c) {
			return c.Next()
		}

		// Defaults to IP
		key := cfg.KeyGenerator(c)

		info, err := store.Take(key)
		if err != nil {
			return failure.Wrap(err, "store.Take failed for (%s)", key)
		}

		reset := time.Unix(0, int64(info.Reset)).UTC().Format(time.RFC1123)

		c.Set(HeaderRateLimitLimit, strconv.FormatUint(info.LimitSize, 10))
		c.Set(HeaderRateLimitRemaining, strconv.FormatUint(info.Remaining, 10))
		c.Set(HeaderRateLimitReset, reset)

		if !info.OperationOk {
			c.Set(HeaderRetryAfter, reset)
			return cfg.Exceeded(c)
		}

		return c.Next()
	}
}
