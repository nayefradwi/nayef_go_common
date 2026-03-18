package pagination

import (
	"net/http"
)

func SetOffsetPaginationMiddleware(f http.Handler) http.Handler {
	hanlder := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pageKey := GetIntQueryParam(r, PageKey)
		pageSizeKey := GetIntQueryParam(r, PageSizeKey)
		query := NewOffsetPageQuery(pageKey, pageSizeKey)

		ctx := query.WithContext(r.Context())
		r = r.WithContext(ctx)

		f.ServeHTTP(w, r)
	})

	return hanlder
}
