package taskuc

import (
	"strings"
	"testing"
	"time"
)

func TestFormatJalaliDateTimeNowruz(t *testing.T) {
	loc := time.FixedZone("IRST", 3*3600+30*60)
	tm := time.Date(2026, 3, 21, 9, 0, 0, 0, loc)
	got := formatJalaliDateTime(tm)
	if !strings.Contains(got, "فروردین") || !strings.Contains(got, "1405") {
		t.Fatalf("got %s", got)
	}
}
