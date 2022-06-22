package limits_test

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"github.com/rsb/api_rate_limiter/foundation/limits"
	"github.com/stretchr/testify/require"
	"sort"
	"testing"
	"time"
)

func TestMemoryStore_FillRate(t *testing.T) {
	t.Parallel()

	t.Run("many_tokens_small_interval", func(t *testing.T) {
		t.Parallel()

		config := limits.Config{
			Limit:    65525,
			Interval: time.Second,
		}
		store := limits.NewMemoryStore(&config)
		go store.GarbageCollector()

		for i := 0; i < 20; i++ {
			info, err := store.Take("key")
			require.NoError(t, err)
			require.False(t, info.Remaining < (info.LimitSize-uint64(i)-1))
			time.Sleep(100 * time.Millisecond)
		}
	})
}

func TestMemoryStore_Take(t *testing.T) {
	cases := []struct {
		name     string
		tokens   uint64
		interval time.Duration
	}{
		{
			name:     "millisecond interval",
			tokens:   5,
			interval: 500 * time.Millisecond,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			key := testKey(t)
			config := limits.Config{
				Interval:    tt.interval,
				Limit:       tt.tokens,
				TTLInterval: 24 * time.Hour,
				MinTTL:      24 * time.Hour,
			}
			store := limits.NewMemoryStore(&config)
			go store.GarbageCollector()

			t.Cleanup(func() {
				err := store.Close()
				require.NoError(t, err)
			})

			type result struct {
				limit, remaining uint64
				reset            time.Duration
				ok               bool
				err              error
			}

			// Take twice everything from the bucket
			takeCH := make(chan *result, 2*tt.tokens)

			for i := uint64(1); i <= 2*tt.tokens; i++ {
				go func() {
					info, err := store.Take(key)
					rs := result{
						limit:     info.LimitSize,
						remaining: info.Remaining,
						reset:     time.Duration(uint64(time.Now().Unix()) - info.Reset),
						ok:        info.OperationOk,
						err:       err,
					}
					takeCH <- &rs
				}()
			}

			// Accumulate and sort results, since they could come in any order.
			var results []*result
			for i := uint64(1); i <= 2*tt.tokens; i++ {
				select {
				case rs := <-takeCH:
					results = append(results, rs)
				case <-time.After(5 * time.Second):
					t.Fatal("timeout")
				}
			}

			sort.Slice(results, func(i, j int) bool {
				if results[i].remaining == results[j].remaining {
					return !results[j].ok
				}
				return results[i].remaining > results[j].remaining
			})

			for i, rs := range results {
				require.NoError(t, rs.err)
				require.Equal(t, rs.limit, tt.tokens)
				require.True(t, rs.reset < tt.interval)

				// first half should pass 2nd half should fail
				if uint64(i) < tt.tokens {
					require.Equal(t, tt.tokens-uint64(i)-1, rs.remaining)
					require.True(t, rs.ok)
				} else {
					require.Equal(t, uint64(0), rs.remaining)
					require.False(t, rs.ok)
				}
			}

			// Wait for the bucket to have entries again
			time.Sleep(tt.interval)

			info, err := store.Take(key)
			require.NoError(t, err)
			require.True(t, info.OperationOk)
		})
	}
}

func testKey(t *testing.T) string {
	t.Helper()

	var b [512]byte
	_, err := rand.Read(b[:])
	require.NoError(t, err)

	digest := fmt.Sprintf("%x", sha256.Sum256(b[:]))
	return digest[:32]
}

func TestMemoryStore_GetBeforeAnySet(t *testing.T) {
	t.Parallel()

	config := limits.Config{
		Limit:       5,
		Interval:    3 * time.Second,
		TTLInterval: 24 * time.Hour,
		MinTTL:      24 * time.Hour,
	}

	store := limits.NewMemoryStore(&config)
	go store.GarbageCollector()

	t.Cleanup(func() {
		err := store.Close()
		require.NoError(t, err)
	})

	key := "my-key"
	limit, remaining, err := store.Get(key)
	require.NoError(t, err)

	require.Equal(t, uint64(0), limit)
	require.Equal(t, uint64(0), remaining)
}

func TestMemoryStore_GetAfterSet(t *testing.T) {
	t.Parallel()

	config := limits.Config{
		Limit:       5,
		Interval:    3 * time.Second,
		TTLInterval: 24 * time.Hour,
		MinTTL:      24 * time.Hour,
	}

	store := limits.NewMemoryStore(&config)
	go store.GarbageCollector()

	t.Cleanup(func() {
		err := store.Close()
		require.NoError(t, err)
	})

	key := "my-key"
	err := store.Set(key, uint64(5), 5*time.Second)
	require.NoError(t, err)

	limit, remaining, err := store.Get(key)
	require.NoError(t, err)

	require.Equal(t, uint64(5), limit)
	require.Equal(t, uint64(5), remaining)
}

func TestMemoryStore_GetAfterTake(t *testing.T) {
	t.Parallel()

	config := limits.Config{
		Limit:       5,
		Interval:    3 * time.Second,
		TTLInterval: 24 * time.Hour,
		MinTTL:      24 * time.Hour,
	}

	store := limits.NewMemoryStore(&config)
	go store.GarbageCollector()

	t.Cleanup(func() {
		err := store.Close()
		require.NoError(t, err)
	})

	key := "my-key"
	info, err := store.Take(key)
	require.NoError(t, err)
	require.Equal(t, uint64(5), info.LimitSize)
	require.Equal(t, uint64(4), info.Remaining)
	require.True(t, info.OperationOk)

	limit, remaining, err := store.Get(key)
	require.NoError(t, err)

	require.Equal(t, uint64(5), limit)
	require.Equal(t, uint64(4), remaining)
}

func TestIntervalCount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		start    uint64
		current  uint64
		interval time.Duration
		expected uint64
	}{
		{
			name:     "count is zero",
			start:    0,
			current:  0,
			interval: time.Second,
			expected: 0,
		},

		{
			name:     "half an interval",
			start:    0,
			current:  uint64(500 * time.Millisecond),
			interval: time.Second,
			expected: 0,
		},
		{
			name:     "1 full interval",
			start:    0,
			current:  uint64(1 * time.Second),
			interval: time.Second,
			expected: 1,
		},
		{
			name:     "not a full interval",
			start:    0,
			current:  uint64(1*time.Second - time.Nanosecond),
			interval: time.Second,
			expected: 0,
		},
		{
			name:     "many intervals with small values",
			start:    0,
			current:  uint64(50*time.Second - 500*time.Millisecond),
			interval: time.Millisecond,
			expected: 49500,
		},
		{
			name:     "several intervals",
			start:    0,
			current:  uint64(100*time.Second - 500*time.Nanosecond),
			interval: time.Second,
			expected: 99,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			count := limits.IntervalCount(tt.start, tt.current, tt.interval)
			require.Equal(t, tt.expected, count)
		})
	}
}
