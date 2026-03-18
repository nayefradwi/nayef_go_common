package pagination

import (
	"net/http"

	"github.com/nayefradwi/nayef_go_common/httputil"
)

func SetOffsetPaginationMiddleware(f http.Handler) http.Handler {
	hanlder := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pageKey := httputil.GetIntQueryParam(r, PageKey)
		pageSizeKey := httputil.GetIntQueryParam(r, PageSizeKey)
		query := NewOffsetPageQuery(pageKey, pageSizeKey)

		ctx := query.WithContext(r.Context())
		r = r.WithContext(ctx)

		f.ServeHTTP(w, r)
	})

	return hanlder
}
