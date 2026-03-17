package rest

import (
	"net/http"
	"strconv"
)

func GetIntQueryParam(r *http.Request, key string) int {
	query := r.URL.Query().Get(key)
	value, _ := strconv.Atoi(query)
	return value
}

func GetBearerToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")

	if authHeader == "" || len(authHeader) < len("Bearer ") {
		return ""
	}

	token := authHeader[len("Bearer "):]
	return token
}
