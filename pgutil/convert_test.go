package pgutil

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func numeric(t *testing.T, s string) pgtype.Numeric {
	t.Helper()
	var n pgtype.Numeric
	if err := n.Scan(s); err != nil {
		t.Fatalf("scan %q: %v", s, err)
	}
	return n
}

func TestFloat64FromNumeric(t *testing.T) {
	if got := Float64FromNumeric(numeric(t, "19.99")); got != 19.99 {
		t.Fatalf("got %v, want 19.99", got)
	}
	if got := Float64FromNumeric(pgtype.Numeric{}); got != 0 {
		t.Fatalf("invalid numeric = %v, want 0", got)
	}
}

func TestTimestamptz(t *testing.T) {
	now := time.Date(2026, 7, 2, 13, 45, 30, 0, time.UTC)
	v := Timestamptz(now)
	if !v.Valid || !v.Time.Equal(now) {
		t.Fatalf("got %+v, want valid %v", v, now)
	}
	if Timestamptz(time.Time{}).Valid {
		t.Fatal("zero time should produce an invalid Timestamptz")
	}

	if got := TimeFromTimestamptz(v); got == nil || !got.Equal(now) {
		t.Fatalf("round trip = %v, want %v", got, now)
	}
	if got := TimeFromTimestamptz(pgtype.Timestamptz{}); got != nil {
		t.Fatalf("invalid timestamptz = %v, want nil", got)
	}
}

func TestDate(t *testing.T) {
	day := time.Date(2026, 7, 2, 0, 0, 0, 0, time.UTC)
	v := Date(day)
	if !v.Valid || !v.Time.Equal(day) {
		t.Fatalf("got %+v, want valid %v", v, day)
	}
	if Date(time.Time{}).Valid {
		t.Fatal("zero time should produce an invalid Date")
	}

	if got := TimeFromDate(v); got == nil || !got.Equal(day) {
		t.Fatalf("round trip = %v, want %v", got, day)
	}
	if got := TimeFromDate(pgtype.Date{}); got != nil {
		t.Fatalf("invalid date = %v, want nil", got)
	}
}

func TestInt16Slices(t *testing.T) {
	days := []int{1, 3, 5}
	got := IntsFromInt16s(Int16s(days))
	if len(got) != len(days) {
		t.Fatalf("got %v, want %v", got, days)
	}
	for i := range days {
		if got[i] != days[i] {
			t.Fatalf("got %v, want %v", got, days)
		}
	}

	if got := Int16s(nil); len(got) != 0 {
		t.Fatalf("Int16s(nil) = %v, want empty", got)
	}
	if got := IntsFromInt16s(nil); len(got) != 0 {
		t.Fatalf("IntsFromInt16s(nil) = %v, want empty", got)
	}
}
