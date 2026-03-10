package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"journal/api/internal/requestmeta"
	"journal/common/errorx"
	commonratelimit "journal/common/ratelimit"
	"journal/common/result"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
)

type limiter interface {
	Allow(ctx context.Context, key string) (bool, int, error)
	Window() time.Duration
}

type routeRule struct {
	method  string
	match   func(path string) bool
	keyFunc func(r *http.Request) string
	limiter limiter
}

type RateLimitMiddleware struct {
	rules []routeRule
}

func NewRateLimitMiddleware(searchLimiter, ratingLimiter, flagLimiter *commonratelimit.TokenBucket) rest.Middleware {
	m := &RateLimitMiddleware{
		rules: []routeRule{
			{
				method:  http.MethodGet,
				match:   matchSearchPath,
				keyFunc: requestIP,
				limiter: searchLimiter,
			},
			{
				method:  http.MethodPost,
				match:   matchPaperRatePath,
				keyFunc: requestUserKey,
				limiter: ratingLimiter,
			},
			{
				method:  http.MethodPost,
				match:   matchFlagPath,
				keyFunc: requestUserKey,
				limiter: flagLimiter,
			},
		},
	}

	return m.Handle
}

func (m *RateLimitMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		for _, rule := range m.rules {
			if rule.method != r.Method || !rule.match(r.URL.Path) || rule.limiter == nil {
				continue
			}

			key := rule.keyFunc(r)
			if key == "" {
				break
			}

			allowed, _, err := rule.limiter.Allow(r.Context(), key)
			if err != nil {
				logx.WithContext(r.Context()).Errorf("rate limit check failed for %s %s: %v", r.Method, r.URL.Path, err)
				break
			}
			if allowed {
				break
			}

			retryAfter := int(rule.limiter.Window().Seconds())
			if retryAfter <= 0 {
				retryAfter = 1
			}

			w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
			httpx.WriteJsonCtx(r.Context(), w, http.StatusTooManyRequests, &result.Response{
				Code:    errorx.ErrRateLimited,
				Message: errorx.CodeMsg(errorx.ErrRateLimited),
			})
			return
		}

		next(w, r)
	}
}

func matchSearchPath(path string) bool {
	return path == "/api/v1/papers/search"
}

func matchPaperRatePath(path string) bool {
	if !strings.HasPrefix(path, "/api/v1/papers/") || !strings.HasSuffix(path, "/rate") {
		return false
	}

	idPart := strings.TrimSuffix(strings.TrimPrefix(path, "/api/v1/papers/"), "/rate")
	return idPart != "" && !strings.Contains(idPart, "/")
}

func matchFlagPath(path string) bool {
	switch {
	case strings.HasPrefix(path, "/api/v1/papers/") && strings.HasSuffix(path, "/flag"):
		idPart := strings.TrimSuffix(strings.TrimPrefix(path, "/api/v1/papers/"), "/flag")
		return idPart != "" && !strings.Contains(idPart, "/")
	case strings.HasPrefix(path, "/api/v1/ratings/") && strings.HasSuffix(path, "/flag"):
		idPart := strings.TrimSuffix(strings.TrimPrefix(path, "/api/v1/ratings/"), "/flag")
		return idPart != "" && !strings.Contains(idPart, "/")
	default:
		return false
	}
}

func requestUserKey(r *http.Request) string {
	if userID, ok := requestUserID(r.Context()); ok {
		return strconv.FormatInt(userID, 10)
	}

	return requestIP(r)
}

func requestUserID(ctx context.Context) (int64, bool) {
	value := ctx.Value("userId")
	switch v := value.(type) {
	case int64:
		return v, true
	case int:
		return int64(v), true
	case float64:
		return int64(v), true
	case json.Number:
		id, err := v.Int64()
		return id, err == nil
	case string:
		id, err := strconv.ParseInt(v, 10, 64)
		return id, err == nil
	default:
		return 0, false
	}
}

func requestIP(r *http.Request) string {
	return requestmeta.RequestIP(r)
}
