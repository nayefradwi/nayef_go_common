package dates

import "time"

func FormatUtcDate() string {
	stripped := StripTimeUTC(time.Now().UTC())
	return TimeToISO8601(stripped)
}

func FormatUtcDateFromTime(t time.Time) string {
	stripped := StripTimeUTC(t)
	return TimeToISO8601(stripped)
}

func StripTime(t time.Time, location *time.Location) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, location)
}

func StripTimeUTC(t time.Time) time.Time {
	return StripTime(t, time.UTC)
}

func TimeToISO8601(t time.Time) string {
	return t.Format(time.RFC3339)
}
