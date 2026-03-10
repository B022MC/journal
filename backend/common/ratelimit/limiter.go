package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

// TokenBucket implements a Redis-based token bucket rate limiter.
// Each bucket is identified by a key and refills at a fixed rate.
type TokenBucket struct {
	store  *redis.Redis
	prefix string
	limit  int
	window time.Duration
}

// NewTokenBucket creates a rate limiter.
// prefix: key prefix (e.g. "rate_limit:rating")
// limit: max requests per window
// window: time window duration
func NewTokenBucket(store *redis.Redis, prefix string, limit int, window time.Duration) *TokenBucket {
	return &TokenBucket{
		store:  store,
		prefix: prefix,
		limit:  limit,
		window: window,
	}
}

// Allow checks whether the request identified by `key` is allowed.
// Returns (allowed bool, remaining int, error).
func (tb *TokenBucket) Allow(ctx context.Context, key string) (bool, int, error) {
	redisKey := fmt.Sprintf("%s:%s", tb.prefix, key)
	windowSec := int(tb.window.Seconds())

	// Use INCR + EXPIRE atomic pattern
	count, err := tb.store.IncrCtx(ctx, redisKey)
	if err != nil {
		return true, 0, err // fail open: allow on Redis error
	}

	// Set expiry only on first request
	if count == 1 {
		_ = tb.store.ExpireCtx(ctx, redisKey, windowSec)
	}

	remaining := tb.limit - int(count)
	if remaining < 0 {
		remaining = 0
	}

	return int(count) <= tb.limit, remaining, nil
}

// Reset clears the rate limit for a specific key
func (tb *TokenBucket) Reset(ctx context.Context, key string) error {
	redisKey := fmt.Sprintf("%s:%s", tb.prefix, key)
	_, err := tb.store.DelCtx(ctx, redisKey)
	return err
}

// Common rate limiter factory functions

// NewRatingLimiter creates a limiter for rating: 20 per hour per user
func NewRatingLimiter(store *redis.Redis) *TokenBucket {
	return NewTokenBucket(store, "rate_limit:rating", 20, time.Hour)
}

// NewFlagLimiter creates a limiter for flagging: 10 per day per user
func NewFlagLimiter(store *redis.Redis) *TokenBucket {
	return NewTokenBucket(store, "rate_limit:flag", 10, 24*time.Hour)
}

// NewSearchLimiter creates a limiter for search: 10 per minute per IP
func NewSearchLimiter(store *redis.Redis) *TokenBucket {
	return NewTokenBucket(store, "rate_limit:search", 10, time.Minute)
}

// Limit returns the maximum allowed requests in the configured window.
func (tb *TokenBucket) Limit() int {
	return tb.limit
}

// Window returns the configured refill window for Retry-After and observability.
func (tb *TokenBucket) Window() time.Duration {
	return tb.window
}
