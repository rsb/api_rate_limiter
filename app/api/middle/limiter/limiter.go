// Package limiter uses the foundation limits package to implement
// rate limiting middleware
package limiter

import (
	"github.com/gofiber/fiber/v2"
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
	Next        func(c *fiber.Ctx) bool
	Limit       uint64
	Interval    time.Duration
	TTLInterval time.Duration
	MinTTL      time.Duration
	StorageSize int
	Exceeded    fiber.Handler
}
