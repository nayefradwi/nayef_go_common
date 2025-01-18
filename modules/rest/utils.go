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
