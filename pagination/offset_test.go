package pagination

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

// NewOffsetPageQuery

func TestNewOffsetPageQuery_Valid(t *testing.T) {
	q := NewOffsetPageQuery(2, 20)
	require.Equal(t, 2, q.Page)
	require.Equal(t, 20, q.PageSize)
}

func TestNewOffsetPageQuery_PageBelowOne_ClampsToOne(t *testing.T) {
	q := NewOffsetPageQuery(0, 10)
	require.Equal(t, 1, q.Page)
}

func TestNewOffsetPageQuery_NegativePage_ClampsToOne(t *testing.T) {
	q := NewOffsetPageQuery(-5, 10)
	require.Equal(t, 1, q.Page)
}

func TestNewOffsetPageQuery_PageSizeBelowOne_DefaultsTen(t *testing.T) {
	q := NewOffsetPageQuery(1, 0)
	require.Equal(t, 10, q.PageSize)
}

func TestNewOffsetPageQuery_PageSizeAbove100_DefaultsTen(t *testing.T) {
	q := NewOffsetPageQuery(1, 200)
	require.Equal(t, 10, q.PageSize)
}

// OffsetPageQuery.Offset

func TestOffset_FirstPage(t *testing.T) {
	q := NewOffsetPageQuery(1, 10)
	require.Equal(t, 0, q.Offset())
}

func TestOffset_SecondPage(t *testing.T) {
	q := NewOffsetPageQuery(2, 10)
	require.Equal(t, 10, q.Offset())
}

func TestOffset_ArbitraryPage(t *testing.T) {
	q := NewOffsetPageQuery(3, 25)
	require.Equal(t, 50, q.Offset())
}

// OffsetPageQueryFromContext

func TestOffsetPageQueryFromContext_Stored(t *testing.T) {
	q := NewOffsetPageQuery(3, 15)
	ctx := q.WithContext(context.Background())
	got := OffsetPageQueryFromContext(ctx)
	require.Equal(t, q, got)
}

func TestOffsetPageQueryFromContext_Empty_ReturnsDefault(t *testing.T) {
	got := OffsetPageQueryFromContext(context.Background())
	require.Equal(t, DefaultPageQuery, got)
}

// NewOffsetPage

func TestNewOffsetPage_FirstPage_MorePages(t *testing.T) {
	page := NewOffsetPage(1, 10, 25, []int{})
	require.True(t, page.HasNext)
	require.False(t, page.HasPrev)
	require.Equal(t, 3, page.NumberOfPages)
}

func TestNewOffsetPage_LastPage(t *testing.T) {
	page := NewOffsetPage(3, 10, 25, []int{})
	require.False(t, page.HasNext)
	require.True(t, page.HasPrev)
	require.Equal(t, 3, page.NumberOfPages)
}

func TestNewOffsetPage_MiddlePage(t *testing.T) {
	page := NewOffsetPage(2, 10, 25, []int{})
	require.True(t, page.HasNext)
	require.True(t, page.HasPrev)
	require.Equal(t, 3, page.NumberOfPages)
}

func TestNewOffsetPage_ExactTotalFit(t *testing.T) {
	page := NewOffsetPage(2, 10, 20, []int{})
	require.False(t, page.HasNext)
	require.True(t, page.HasPrev)
	require.Equal(t, 2, page.NumberOfPages)
}

func TestNewOffsetPage_SinglePage(t *testing.T) {
	page := NewOffsetPage(1, 10, 5, []int{})
	require.False(t, page.HasNext)
	require.False(t, page.HasPrev)
	require.Equal(t, 1, page.NumberOfPages)
}

func TestNewOffsetPage_ItemsPopulated(t *testing.T) {
	items := []string{"a", "b"}
	page := NewOffsetPage(1, 2, 3, items)
	require.Equal(t, items, page.Items)
}

// NewEmptyPage

func TestNewEmptyPage(t *testing.T) {
	page := NewEmptyPage[int](10)
	require.Equal(t, 1, page.Page)
	require.False(t, page.HasNext)
	require.False(t, page.HasPrev)
	require.Equal(t, 1, page.NumberOfPages)
	require.Equal(t, 0, page.TotalItems)
	require.Empty(t, page.Items)
}

// SetOffsetPaginationMiddleware

func newMiddlewareHandler(t *testing.T, assertFn func(q OffsetPageQuery)) http.Handler {
	t.Helper()
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertFn(OffsetPageQueryFromContext(r.Context()))
		w.WriteHeader(http.StatusOK)
	})
	return SetOffsetPaginationMiddleware(next)
}

func TestSetOffsetPaginationMiddleware_ValidParams(t *testing.T) {
	handler := newMiddlewareHandler(t, func(q OffsetPageQuery) {
		require.Equal(t, 2, q.Page)
		require.Equal(t, 20, q.PageSize)
	})
	req := httptest.NewRequest(http.MethodGet, "/?page=2&page_size=20", nil)
	handler.ServeHTTP(httptest.NewRecorder(), req)
}

func TestSetOffsetPaginationMiddleware_MissingParams_UseDefaults(t *testing.T) {
	handler := newMiddlewareHandler(t, func(q OffsetPageQuery) {
		require.Equal(t, 1, q.Page)
		require.Equal(t, 10, q.PageSize)
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(httptest.NewRecorder(), req)
}

func TestSetOffsetPaginationMiddleware_InvalidParams_UseDefaults(t *testing.T) {
	handler := newMiddlewareHandler(t, func(q OffsetPageQuery) {
		require.Equal(t, 1, q.Page)
		require.Equal(t, 10, q.PageSize)
	})
	req := httptest.NewRequest(http.MethodGet, "/?page=abc&page_size=xyz", nil)
	handler.ServeHTTP(httptest.NewRecorder(), req)
}

func TestSetOffsetPaginationMiddleware_OutOfRangePageSize_UseDefaults(t *testing.T) {
	handler := newMiddlewareHandler(t, func(q OffsetPageQuery) {
		require.Equal(t, 1, q.Page)
		require.Equal(t, 10, q.PageSize)
	})
	req := httptest.NewRequest(http.MethodGet, "/?page=1&page_size=200", nil)
	handler.ServeHTTP(httptest.NewRecorder(), req)
}
