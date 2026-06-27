package pagination

import (
	"net/http"

	"github.com/nayefradwi/nayef_go_common/httputil"
)

func SetCursorPaginationMiddleware(f http.Handler) http.Handler {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		after := r.URL.Query().Get(AfterKey)
		before := r.URL.Query().Get(BeforeKey)
		pageSize := httputil.GetIntQueryParam(r, PageSizeKey)
		query := NewCursorPageQuery(after, before, pageSize)

		ctx := query.WithContext(r.Context())
		r = r.WithContext(ctx)

		f.ServeHTTP(w, r)
	})

	return handler
}
