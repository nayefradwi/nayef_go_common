package shared

import (
	"cmp"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/nayefradwi/nayef_go_common/errors"
)

var ErrTooManyRequests = errors.NewResultErrorWithStatus("Too many requests", "TOO_MANY_REQUESTS", http.StatusTooManyRequests)

func Must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}

	return v
}

func Decode[T any](r *http.Request) (T, error) {
	var body T
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return body, errors.BadRequestError("invalid json body")
	}

	return body, nil
}

// lowercased so "login:"+email style keys and the users table agree on one spelling
func ParseEmail(raw string) (string, error) {
	email := strings.ToLower(strings.TrimSpace(raw))
	if !strings.Contains(email, "@") || strings.ContainsAny(email, " \t\n") {
		return "", errors.NewValidationError(errors.Field("email", "invalid", "email is invalid"))
	}

	return email, nil
}

// Serve blocks until the server stops; ADDR defaults to :8080.
func Serve(h http.Handler) {
	addr := cmp.Or(os.Getenv("ADDR"), ":8080")
	slog.Info("listening", "addr", addr)
	slog.Error("server stopped", "error", http.ListenAndServe(addr, h))
}
