package rest

import (
	"net/http"

	"github.com/nayefradwi/nayef_go_common/modules/pagination"
	"github.com/nayefradwi/nayef_go_common/modules/rest"
)

func SetOffsetPaginationMiddleware(f http.HandlerFunc) http.HandlerFunc {
	hanlder := func(w http.ResponseWriter, r *http.Request) {
		pageKey := rest.GetIntQueryParam(r, pagination.PageKey)
		pageSizeKey := rest.GetIntQueryParam(r, pagination.PageSizeKey)
		query := pagination.NewOffsetPageQuery(pageKey, pageSizeKey)

		ctx := query.WithContext(r.Context())
		r = r.WithContext(ctx)

		f(w, r)
	}

	return hanlder
}
