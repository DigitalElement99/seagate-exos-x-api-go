package common

import "testing"

// Regression (2026-10-23): creation-date-time is the array's local wall clock
// with no zone ("2026-10-23 12:09:14" on a -07:00 array); parsed as UTC it put
// every snapshot 7 h early. creation-date-time-numeric is the true epoch.
func TestCreationTimePrefersEpoch(t *testing.T) {
	ts, err := CreationTime(1792782554, "2026-10-23 12:09:14")
	if err != nil {
		t.Fatal(err)
	}
	if got := ts.AsTime().UTC().Format("2006-01-02 15:04:05"); got != "2026-10-23 19:09:14" {
		t.Errorf("epoch: got %s, want 2026-10-23 19:09:14", got)
	}
	ts, err = CreationTime(0, "2026-10-23 12:09:14")
	if err != nil || ts.AsTime().UTC().Format("15:04:05") != "12:09:14" {
		t.Errorf("fallback to the string when no epoch: got %v, %v", ts, err)
	}
}
