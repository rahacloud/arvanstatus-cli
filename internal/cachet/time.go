package cachet

import "time"

// tehranLoc is the status page's timezone. Cachet emits naive datetimes (no Z,
// no offset) rendered in its configured local time; for ArvanCloud that is
// Asia/Tehran. time.Parse would otherwise silently interpret them as UTC.
var tehranLoc = func() *time.Location {
	loc, err := time.LoadLocation("Asia/Tehran")
	if err != nil {
		return time.FixedZone("IRST", int((3*time.Hour + 30*time.Minute).Seconds()))
	}
	return loc
}()

// cachetLayout matches the Cachet timestamp `string` field, e.g.
// "2026-07-08 19:00:47".
const cachetLayout = "2006-01-02 15:04:05"

// LocalTime wraps time.Time so callers can rely on the same helpers regardless
// of the on-the-wire format. The zero value reports IsZero() == true.
type LocalTime struct {
	time.Time
}

// parseStamp turns a Cachet stamp into a LocalTime, interpreting the naive
// datetime in Asia/Tehran. An empty or unparseable string yields the zero time.
func parseStamp(s stamp) LocalTime {
	if s.String == "" {
		return LocalTime{}
	}
	t, err := time.ParseInLocation(cachetLayout, s.String, tehranLoc)
	if err != nil {
		return LocalTime{}
	}
	return LocalTime{Time: t}
}
