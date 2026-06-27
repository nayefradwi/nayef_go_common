package pagination

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type testCursor struct {
	CreatedAt time.Time `json:"c"`
	ID        int64     `json:"i"`
}

// EncodeCursor / DecodeCursor

func TestCursor_RoundTrip_PreservesTypedFields(t *testing.T) {
	in := testCursor{CreatedAt: time.Unix(1700000000, 0).UTC(), ID: 9007199254740993}
	out, has, err := DecodeCursor[testCursor](EncodeCursor(in))
	require.NoError(t, err)
	require.True(t, has)
	require.True(t, in.CreatedAt.Equal(out.CreatedAt))
	require.Equal(t, in.ID, out.ID)
}

func TestDecodeCursor_Empty_NoCursor(t *testing.T) {
	out, has, err := DecodeCursor[testCursor]("")
	require.NoError(t, err)
	require.False(t, has)
	require.Equal(t, testCursor{}, out)
}

func TestDecodeCursor_BadBase64_Errors(t *testing.T) {
	_, has, err := DecodeCursor[testCursor]("not!valid!base64")
	require.Error(t, err)
	require.False(t, has)
}

// NewCursorPageQuery

func TestNewCursorPageQuery_Valid(t *testing.T) {
	q := NewCursorPageQuery("a", "", 20)
	require.Equal(t, "a", q.After)
	require.Equal(t, 20, q.PageSize)
}

func TestNewCursorPageQuery_PageSizeBelowOne_DefaultsTen(t *testing.T) {
	require.Equal(t, 10, NewCursorPageQuery("", "", 0).PageSize)
}

func TestNewCursorPageQuery_PageSizeAbove100_DefaultsTen(t *testing.T) {
	require.Equal(t, 10, NewCursorPageQuery("", "", 200).PageSize)
}

func TestCursorPageQuery_ForwardAndCursor(t *testing.T) {
	fwd := NewCursorPageQuery("a", "", 10)
	require.True(t, fwd.Forward())
	require.Equal(t, "a", fwd.Cursor())

	bwd := NewCursorPageQuery("a", "b", 10)
	require.False(t, bwd.Forward())
	require.Equal(t, "b", bwd.Cursor())
}

func TestCursorPageQueryFromContext_Stored(t *testing.T) {
	q := NewCursorPageQuery("a", "b", 15)
	got := CursorPageQueryFromContext(q.WithContext(context.Background()))
	require.Equal(t, q, got)
}

func TestCursorPageQueryFromContext_Empty_ReturnsDefault(t *testing.T) {
	got := CursorPageQueryFromContext(context.Background())
	require.Equal(t, NewCursorPageQuery("", "", 10), got)
}

// CursorBuilder.Execute

func rowsOf(n int) []int {
	out := make([]int, n)
	for i := range out {
		out[i] = i
	}
	return out
}

func forwardReturning(rows []int) func(context.Context, testCursor, bool, int32) ([]int, error) {
	return func(context.Context, testCursor, bool, int32) ([]int, error) { return rows, nil }
}

func keyInt(v int) testCursor { return testCursor{ID: int64(v)} }

func TestExecute_Forward_FullPageWithExtra_HasNext(t *testing.T) {
	q := NewCursorPageQuery("", "", 3)
	page, err := NewCursor[int, testCursor](q).
		OnForward(forwardReturning(rowsOf(4))). // PageSize+1
		Key(keyInt).
		Execute(context.Background())
	require.NoError(t, err)
	require.True(t, page.HasNext)
	require.False(t, page.HasPrev)
	require.Len(t, page.Items, 3) // probe trimmed
}

func TestExecute_Forward_ExactFit_NoNext(t *testing.T) {
	q := NewCursorPageQuery("", "", 3)
	page, err := NewCursor[int, testCursor](q).
		OnForward(forwardReturning(rowsOf(3))). // exactly PageSize
		Key(keyInt).
		Execute(context.Background())
	require.NoError(t, err)
	require.False(t, page.HasNext)
	require.Len(t, page.Items, 3)
}

func TestExecute_Forward_WithCursor_HasPrev(t *testing.T) {
	q := NewCursorPageQuery(EncodeCursor(testCursor{ID: 5}), "", 3)
	page, err := NewCursor[int, testCursor](q).
		OnForward(forwardReturning(rowsOf(2))).
		Key(keyInt).
		Execute(context.Background())
	require.NoError(t, err)
	require.True(t, page.HasPrev)
}

func TestExecute_Backward_ReversesAndFlags(t *testing.T) {
	q := NewCursorPageQuery("", EncodeCursor(testCursor{ID: 9}), 3)
	page, err := NewCursor[int, testCursor](q).
		OnBackward(func(context.Context, testCursor, int32) ([]int, error) {
			return []int{0, 1, 2, 3}, nil // ascending query order, with extra
		}).
		Key(keyInt).
		Execute(context.Background())
	require.NoError(t, err)
	require.True(t, page.HasNext)
	require.True(t, page.HasPrev)                // extra present
	require.Equal(t, []int{2, 1, 0}, page.Items) // probe (tail) trimmed, then reversed
}

func TestExecute_Backward_NoReverse_KeepsOrder(t *testing.T) {
	q := NewCursorPageQuery("", EncodeCursor(testCursor{ID: 9}), 3)
	page, err := NewCursor[int, testCursor](q).
		OnBackward(func(context.Context, testCursor, int32) ([]int, error) {
			return []int{2, 1, 0}, nil
		}).
		Key(keyInt).
		NoReverse().
		Execute(context.Background())
	require.NoError(t, err)
	require.Equal(t, []int{2, 1, 0}, page.Items)
}

func TestExecute_Backward_Unsupported_Errors(t *testing.T) {
	q := NewCursorPageQuery("", EncodeCursor(testCursor{ID: 9}), 3)
	_, err := NewCursor[int, testCursor](q).
		OnForward(forwardReturning(rowsOf(1))).
		Key(keyInt).
		Execute(context.Background())
	require.ErrorIs(t, err, ErrBackwardUnsupported)
}

func TestExecute_EmptyResult_EmptyCursors(t *testing.T) {
	q := NewCursorPageQuery("", "", 3)
	page, err := NewCursor[int, testCursor](q).
		OnForward(forwardReturning([]int{})).
		Key(keyInt).
		Execute(context.Background())
	require.NoError(t, err)
	require.False(t, page.HasNext)
	require.Empty(t, page.After)
	require.Empty(t, page.Before)
}

// MapCursorPage

func TestMapCursorPage_TransformsItems_PreservesMeta(t *testing.T) {
	in := NewCursorPage(true, true, []int{1, 2}, "a", "b")
	out := MapCursorPage(in, func(v int) int { return v * 10 })
	require.Equal(t, []int{10, 20}, out.Items)
	require.True(t, out.HasNext)
	require.True(t, out.HasPrev)
	require.Equal(t, "a", out.After)
	require.Equal(t, "b", out.Before)
}

// SetCursorPaginationMiddleware

func newCursorMiddlewareHandler(t *testing.T, assertFn func(q CursorPageQuery)) http.Handler {
	t.Helper()
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertFn(CursorPageQueryFromContext(r.Context()))
		w.WriteHeader(http.StatusOK)
	})
	return SetCursorPaginationMiddleware(next)
}

func TestSetCursorPaginationMiddleware_ValidParams(t *testing.T) {
	handler := newCursorMiddlewareHandler(t, func(q CursorPageQuery) {
		require.Equal(t, "x", q.After)
		require.Equal(t, "y", q.Before)
		require.Equal(t, 20, q.PageSize)
	})
	req := httptest.NewRequest(http.MethodGet, "/?after=x&before=y&page_size=20", nil)
	handler.ServeHTTP(httptest.NewRecorder(), req)
}

func TestSetCursorPaginationMiddleware_MissingParams_UseDefaults(t *testing.T) {
	handler := newCursorMiddlewareHandler(t, func(q CursorPageQuery) {
		require.Empty(t, q.After)
		require.Empty(t, q.Before)
		require.Equal(t, 10, q.PageSize)
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(httptest.NewRecorder(), req)
}
