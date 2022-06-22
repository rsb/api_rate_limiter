// Package limits is an example implementation of a rate limits that
// focuses on Fixed time window rate limit
package limits

import (
	"github.com/rsb/failure"
	"sync"
	"sync/atomic"
	"time"
)

const (
	DefaultRateLimit         = uint64(1)
	DefaultRateLimitInterval = 1 * time.Second
	DefaultTTLInterval       = 6 * time.Hour
	DefaultMinTTLInterval    = 12 * time.Hour
	DefaultInitialMapSize    = 4096
)

type RateInfo struct {
	LimitSize   uint64
	Remaining   uint64
	Reset       uint64
	OperationOk bool
}

// Config controls the configuration of the storage systems
//
// Limit 			 - the number of limit allowed per interval default is 1.
// Interval 	 - rate limit interval default is 1 sec.
// TTLInterval - rate at which to clean stale entries. default is 6 hours
// MinTTL      - the minimum amount of time a session must be inactive before
// 							 clearing it from the entries.
// InitialSize - the size to use for map. Go will automatically
// 							 expand the buffer. The default is 4096
type Config struct {
	Limit       uint64
	Interval    time.Duration
	TTLInterval time.Duration
	MinTTL      time.Duration
	InitialSize int
}

func NewDefaultConfig() *Config {
	return &Config{
		Limit:       DefaultRateLimit,
		Interval:    DefaultRateLimitInterval,
		TTLInterval: DefaultTTLInterval,
		MinTTL:      DefaultMinTTLInterval,
		InitialSize: DefaultInitialMapSize,
	}
}

// TTL holds values used to control data garbage collection
type TTL struct {
	Interval time.Duration
	Value    uint64
}

func NewTTL(interval time.Duration, purgeTime uint64) TTL {
	return TTL{
		Interval: interval,
		Value:    purgeTime,
	}
}

type MemoryStore struct {
	limit    uint64
	interval time.Duration

	ttl TTL

	data map[string]*Bucket
	lock sync.RWMutex

	stopped uint32
	stop    chan struct{}
}

// NewMemoryStore is the main constructor used to create and configure the
// in-memory storage
func NewMemoryStore(opts ...*Config) *MemoryStore {
	var config *Config
	defaults := NewDefaultConfig()
	if len(opts) > 0 && opts[0] != nil {
		config = opts[0]
	} else {
		config = defaults
	}

	tokens := defaults.Limit
	if config.Limit > 0 {
		tokens = config.Limit
	}

	interval := defaults.Interval
	if config.Interval > 0 {
		interval = config.Interval
	}

	sweepInterval := defaults.TTLInterval
	if config.TTLInterval > 0 {
		sweepInterval = config.TTLInterval
	}

	sweepMinTTL := defaults.MinTTL
	if config.MinTTL > 0 {
		sweepMinTTL = config.MinTTL
	}

	size := defaults.InitialSize
	if config.InitialSize > 0 {
		size = config.InitialSize
	}

	store := MemoryStore{
		limit:    tokens,
		interval: interval,
		ttl:      NewTTL(sweepInterval, uint64(sweepMinTTL)),
		data:     make(map[string]*Bucket, size),
		stop:     make(chan struct{}),
	}

	return &store
}

func (m *MemoryStore) Take(key string) (RateInfo, error) {
	var info RateInfo
	if atomic.LoadUint32(&m.stopped) == 1 {
		return info, failure.InvalidState("MemoryStore is stopped")
	}

	// Acquire a read lock first - this allows others to concurrently check limits
	// without full locks
	m.lock.RLock()
	if b, ok := m.data[key]; ok {
		m.lock.RUnlock()
		return b.RateInfo(), nil
	}
	m.lock.RUnlock()

	// Did not find the key in the map. Take out a full lock. We have
	// to check if the key exists again, because its possible another
	// goroutine created it between our shared lock and exclusive lock
	m.lock.Lock()
	if b, ok := m.data[key]; ok {
		m.lock.Unlock()
		return b.RateInfo(), nil
	}

	// This is a new entry. so create the bucket and take an initial request
	b := NewBucket(m.limit, m.interval)
	m.data[key] = b
	m.lock.Unlock()

	return b.RateInfo(), nil
}

func (m *MemoryStore) Get(key string) (uint64, uint64, error) {
	var tokens, remaining uint64
	if atomic.LoadUint32(&m.stopped) == 1 {
		return tokens, remaining, failure.InvalidState("MemoryStore is stopped")
	}

	m.lock.RLock()
	if b, ok := m.data[key]; ok {
		m.lock.RUnlock()
		tokens, remaining = b.Get()
		return tokens, remaining, nil
	}
	m.lock.RUnlock()

	return tokens, remaining, nil
}

func (m *MemoryStore) Set(key string, tokens uint64, interval time.Duration) error {
	m.lock.Lock()
	b := NewBucket(tokens, interval)
	m.data[key] = b
	m.lock.Unlock()
	return nil
}

// Burst adds the provided value to the bucket's currently available limit
func (m *MemoryStore) Burst(key string, tokens uint64) error {
	m.lock.Lock()
	if b, ok := m.data[key]; ok {
		b.lock.Lock()
		m.lock.Unlock()
		b.availableTokens = b.availableTokens + tokens
		b.lock.Unlock()
		return nil
	}

	// this is a new record for the key
	b := NewBucket(m.limit+tokens, m.interval)
	m.data[key] = b
	m.lock.Unlock()
	return nil
}

// Close stops the memory limits and cleans up any outstanding sessions
// You should always call this method as it releases the memory consumed
// by the map and releases the tickets.
func (m *MemoryStore) Close() error {
	if !atomic.CompareAndSwapUint32(&m.stopped, 0, 1) {
		return nil
	}

	// Close the channel to prevent future purging
	close(m.stop)

	// Delete all data
	m.lock.Lock()
	for key := range m.data {
		delete(m.data, key)
	}
	m.lock.Unlock()
	return nil
}

// GarbageCollector continually iterates over the map and purges old values on the provided
// sweep interval.
func (m *MemoryStore) GarbageCollector() {
	ticker := time.NewTicker(m.ttl.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-m.stop:
			return
		case <-ticker.C:
		}

		m.lock.Lock()
		now := uint64(time.Now().UnixNano())
		for k, b := range m.data {
			b.lock.Lock()
			lastTime := b.startTime + (b.lastTick * uint64(b.interval))
			b.lock.Unlock()

			if now-lastTime > m.ttl.Value {
				delete(m.data, k)
			}
		}
		m.lock.Unlock()
	}
}

// Bucket holds metadata about the rate limit for a given key
//
// startTime 				- the number of nanoseconds from unix epoch when the bucket was created
// maxToken  				- the max number of limit permitted on the bucket at any time. the
//             			  number of available limit will never exceed this value.
// interval  				- the time at which a tick should occur
// availableTokens 	- current number of available limit
// lastTick  				- the last clock tick. used to re-calculate the number of limit on the bucket
// lock 					  - mutex lock to guard the struct fields
type Bucket struct {
	startTime       uint64
	maxTokens       uint64
	interval        time.Duration
	availableTokens uint64
	lastTick        uint64
	lock            sync.Mutex
}

func NewBucket(tokens uint64, interval time.Duration) *Bucket {
	return &Bucket{
		startTime:       uint64(time.Now().UnixNano()),
		maxTokens:       tokens,
		availableTokens: tokens,
		interval:        interval,
	}
}

func (b *Bucket) Get() (uint64, uint64) {
	b.lock.Lock()
	defer b.lock.Unlock()

	return b.maxTokens, b.availableTokens
}

func (b *Bucket) RateInfo() RateInfo {
	var tokens uint64
	var remaining uint64
	var reset uint64
	var ok bool

	now := uint64(time.Now().UnixNano())
	currentTick := IntervalCount(b.startTime, now, b.interval)

	tokens = b.maxTokens
	reset = b.startTime + ((currentTick + 1) * uint64(b.interval))

	b.lock.Lock()
	defer b.lock.Unlock()

	// If we're on a new tick since last assessment, perform a full reset up to maxTokens
	if b.lastTick < currentTick {
		b.availableTokens = b.maxTokens
		b.lastTick = currentTick
	}

	if b.availableTokens > 0 {
		b.availableTokens--
		ok = true
		remaining = b.availableTokens
	}

	return RateInfo{
		LimitSize:   tokens,
		Remaining:   remaining,
		Reset:       reset,
		OperationOk: ok,
	}
}

// IntervalCount is the total number times the current interval has occurred between
// when the time started (start) and the current time (current). For example,
// if the start time was 12:30pm and its current 1:00pm, and the interval was
// 5 minutes, tick would return 6 because 1:00pm is the 6th 5-minute tick.
func IntervalCount(start, current uint64, interval time.Duration) uint64 {
	return (current - start) / uint64(interval.Nanoseconds())
}
