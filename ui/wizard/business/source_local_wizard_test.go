package business

import (
	"testing"

	"singbox-launcher/core/config"
)

func TestLocalAutoOutboundTag(t *testing.T) {
	if g := LocalAutoOutboundTag("1:", 0); g != "1:auto" {
		t.Fatalf("got %q", g)
	}
	if g := LocalAutoOutboundTag("", 2); g != "3:auto" {
		t.Fatalf("empty prefix should use index+1, got %q", g)
	}
}

func TestEnsureLocalAutoSelect(t *testing.T) {
	ps := &config.ProxySource{TagPrefix: "2:"}
	if err := EnsureLocalAuto(ps, 1); err != nil {
		t.Fatal(err)
	}
	if !ProxyHasLocalAuto(ps) {
		t.Fatal("expected auto")
	}
	if err := EnsureLocalSelect(ps, 1); err != nil {
		t.Fatal(err)
	}
	if !ProxyHasLocalSelect(ps) {
		t.Fatal("expected select")
	}
	if len(ps.Outbounds) < 2 {
		t.Fatalf("outbounds %d", len(ps.Outbounds))
	}
}
