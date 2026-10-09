package authhttp

import (
	"log/slog"
	"math"
	"net"
	"net/http"
	"strconv"

	"github.com/nayefradwi/nayef_go_common/auth"
	"github.com/nayefradwi/nayef_go_common/errors"
	"github.com/nayefradwi/nayef_go_common/httputil"
)

type KeyFunc func(r *http.Request) (string, error)

func RateLimit(limiter auth.Limiter, key KeyFunc) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			jw := httputil.NewJsonResponseWriter(w)
			k, err := key(r)
			if err != nil {
				jw.WriteError(err)
				return
			}

			allowed, retryAfter, err := limiter.Allow(r.Context(), k)
			if err != nil {
				slog.ErrorContext(r.Context(), "rate limit check failed", "error", err)
				jw.WriteError(errors.InternalError("failed to check rate limit"))
				return
			}
			if !allowed {
				jw.SetHeader("Retry-After", strconv.Itoa(int(math.Ceil(retryAfter.Seconds()))))
				jw.WriteError(errors.NewResultErrorWithStatus("Too many requests", "TOO_MANY_REQUESTS", http.StatusTooManyRequests))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Uses r.RemoteAddr; behind a proxy, mount a real-IP middleware first.
func ByIP(prefix string) KeyFunc {
	return func(r *http.Request) (string, error) {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			host = r.RemoteAddr
		}
		return prefix + ":" + host, nil
	}
}
