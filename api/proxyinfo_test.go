package api

import "testing"

func TestProxyInfo_ContextMenuTypeLine(t *testing.T) {
	const unk = "?"
	if got := (ProxyInfo{}).ContextMenuTypeLine(unk); got != unk {
		t.Errorf("empty: got %q want %q", got, unk)
	}
	if got := (ProxyInfo{ClashType: "  VLESS  "}).ContextMenuTypeLine(unk); got != "vless" {
		t.Errorf("vless: got %q", got)
	}
	if got := (ProxyInfo{ClashType: "Selector"}).ContextMenuTypeLine(unk); got != "selector" {
		t.Errorf("selector: got %q", got)
	}
}
