package repository

import "testing"

func TestCoalesceHistorySnapshot(t *testing.T) {
	primary := "snapshot-value"
	fallback := "live-value"

	if got := coalesceHistorySnapshot(&primary, &fallback, false); got == nil || *got != primary {
		t.Fatalf("expected primary snapshot to win when present")
	}

	if got := coalesceHistorySnapshot(nil, &fallback, false); got != nil {
		t.Fatalf("expected nil when snapshot is missing and legacy fallback is disabled")
	}

	if got := coalesceHistorySnapshot(nil, &fallback, true); got == nil || *got != fallback {
		t.Fatalf("expected fallback when legacy fallback is enabled")
	}
}

