package api

import (
	"net/url"
	"strings"
	"testing"
)

func TestProxyNamePathEscape_noRawSpaces(t *testing.T) {
	name := "abvpn:🇱🇹PRO > Литва#4 > XRay abvpn"
	enc := url.PathEscape(name)
	if strings.Contains(enc, " ") {
		t.Fatalf("path segment must not contain raw spaces: %q", enc)
	}
	// round-trip
	dec, err := url.PathUnescape(enc)
	if err != nil || dec != name {
		t.Fatalf("PathUnescape: err=%v dec=%q want %q", err, dec, name)
	}
}

func TestPingTestAllConcurrency_allowedValues(t *testing.T) {
	prev := pingTestAllConcurrency
	t.Cleanup(func() { pingTestAllConcurrency = prev })

	for _, want := range []int{1, 5, 10, 20} {
		SetPingTestAllConcurrency(want)
		if got := GetPingTestAllConcurrency(); got != want {
			t.Fatalf("Set %d: got %d", want, got)
		}
	}
	SetPingTestAllConcurrency(7)
	if got := GetPingTestAllConcurrency(); got != 20 {
		t.Fatalf("invalid value: want 20, got %d", got)
	}
}
