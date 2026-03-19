package httputil

import (
	"net/http"
	"strconv"
	"strings"
)

func GetIntQueryParam(r *http.Request, key string) int {
	query := r.URL.Query().Get(key)
	value, _ := strconv.Atoi(query)
	return value
}

func GetBearerToken(r *http.Request) string {
	header := r.Header.Get("Authorization")
	if !strings.HasPrefix(header, "Bearer ") {
		return ""
	}
	return strings.TrimPrefix(header, "Bearer ")
}
