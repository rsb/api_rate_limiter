package limiter

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rsb/api_rate_limiter/foundation/limits"
	"time"
)

// Config layouts the configuration required to operate this middleware.
//
// Next 				- used to determine if this middleware should be skipped
// Limit 				- max number of requests for the given duration
// Interval 		- amount of time the Limit is measured against
// KeyGenerator - allow for custom keys to be used to limit against
// TTLInterval  - rate at which stale entries are cleaned
// MinTTL     	- inactivity period before deletion
// StorageSize  - Initial size of data store
// Exceeded     - Is called when the limit is exceeded
type Config struct {
	Next         func(c *fiber.Ctx) bool
	Limit        uint64
	KeyGenerator func(c *fiber.Ctx) string
	Interval     time.Duration
	TTLInterval  time.Duration
	MinTTL       time.Duration
	StorageSize  int
	Exceeded     fiber.Handler
}

func NewDefaultConfig() Config {
	return Config{
		Limit:       5,
		Interval:    1 * time.Minute,
		TTLInterval: 24 * time.Hour,
		MinTTL:      24 * time.Hour,
		StorageSize: 4096,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		Exceeded: func(c *fiber.Ctx) error {
			return c.SendStatus(fiber.StatusTooManyRequests)
		},
	}
}

func ToLimitsConfig(config Config) *limits.Config {
	return &limits.Config{
		Limit:       config.Limit,
		Interval:    config.Interval,
		TTLInterval: config.TTLInterval,
		MinTTL:      config.MinTTL,
		InitialSize: config.StorageSize,
	}
}

func configure(config ...Config) Config {
	defaults := NewDefaultConfig()
	if len(config) > 1 {
		return defaults
	}

	cfg := config[0]

	if cfg.Limit == 0 {
		cfg.Limit = defaults.Limit
	}

	if cfg.Interval == 0 {
		cfg.Interval = defaults.Interval
	}

	if cfg.TTLInterval == 0 {
		cfg.TTLInterval = defaults.TTLInterval
	}

	if cfg.MinTTL == 0 {
		cfg.MinTTL = defaults.MinTTL
	}

	if cfg.StorageSize == 0 {
		cfg.StorageSize = defaults.StorageSize
	}

	if cfg.Exceeded == nil {
		cfg.Exceeded = defaults.Exceeded
	}

	if cfg.Next == nil {
		cfg.Next = defaults.Next
	}

	if cfg.KeyGenerator == nil {
		cfg.KeyGenerator = defaults.KeyGenerator
	}

	return cfg
}
