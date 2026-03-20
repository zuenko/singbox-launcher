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
