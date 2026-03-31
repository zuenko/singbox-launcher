package textnorm

import "testing"

func TestNormalizeProxyDisplay(t *testing.T) {
	in := "abvpn:🇱🇹PRO ❯ Литва#4 ❯ XRay abvpn"
	want := "abvpn:🇱🇹PRO > Литва#4 > XRay abvpn"
	got := NormalizeProxyDisplay(in)
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	if NormalizeProxyDisplay("") != "" {
		t.Fatal("empty in should be empty out")
	}
}
