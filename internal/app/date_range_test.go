package app

import (
	"testing"
	"time"
)

func TestResolveDateRangeAcceptsSingleDigitMonthDay(t *testing.T) {
	opts := Options{
		Date: "2026-1-5",
	}

	got, err := resolveDateRange(opts, func() time.Time {
		return time.Date(2026, 1, 10, 0, 0, 0, 0, time.Local)
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantStart := time.Date(2026, 1, 5, 0, 0, 0, 0, time.Local)
	wantEnd := wantStart.AddDate(0, 0, 1)
	if !got.Start.Equal(wantStart) {
		t.Fatalf("unexpected start: %v", got.Start)
	}
	if !got.End.Equal(wantEnd) {
		t.Fatalf("unexpected end: %v", got.End)
	}
	if got.IsRange {
		t.Fatalf("expected IsRange=false")
	}
}

func TestResolveDateRangeAcceptsZeroPaddedDate(t *testing.T) {
	opts := Options{
		Date: "2026-01-05",
	}

	_, err := resolveDateRange(opts, func() time.Time {
		return time.Date(2026, 1, 10, 0, 0, 0, 0, time.Local)
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestResolveDateRangeAcceptsSingleDigitRange(t *testing.T) {
	opts := Options{
		From: "2026-1-5",
		To:   "2026-1-7",
	}

	got, err := resolveDateRange(opts, func() time.Time {
		return time.Date(2026, 1, 10, 0, 0, 0, 0, time.Local)
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantStart := time.Date(2026, 1, 5, 0, 0, 0, 0, time.Local)
	wantEnd := time.Date(2026, 1, 8, 0, 0, 0, 0, time.Local)
	if !got.Start.Equal(wantStart) {
		t.Fatalf("unexpected start: %v", got.Start)
	}
	if !got.End.Equal(wantEnd) {
		t.Fatalf("unexpected end: %v", got.End)
	}
	if !got.IsRange {
		t.Fatalf("expected IsRange=true")
	}
}

func TestResolveDateRangeRejectsMissingYear(t *testing.T) {
	opts := Options{
		Date: "1-5",
	}

	_, err := resolveDateRange(opts, func() time.Time {
		return time.Date(2026, 1, 10, 0, 0, 0, 0, time.Local)
	})
	if err == nil {
		t.Fatalf("expected error")
	}
}
