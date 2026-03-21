package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestShareProxyURIForOutboundTag(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "config.json")
	body := []byte(`{"outbounds":[{"type":"vless","tag":"n1","uuid":"550e8400-e29b-41d4-a716-446655440000","server":"example.com","server_port":443,"tls":{"enabled":true,"server_name":"example.com"}}]}`)
	if err := os.WriteFile(p, body, 0644); err != nil {
		t.Fatal(err)
	}
	u, err := ShareProxyURIForOutboundTag(p, "n1")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(u, "vless://") {
		t.Fatalf("expected vless URI, got %q", u)
	}
}

func TestGetOutboundMapByTag_NotFound(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "config.json")
	if err := os.WriteFile(p, []byte(`{"outbounds":[]}`), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := GetOutboundMapByTag(p, "missing")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestShareProxyURIForOutboundTag_WireGuardEndpoint(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "config.json")
	body := []byte(`{"endpoints":[{"type":"wireguard","tag":"wg1","name":"singbox-wg0","system":false,"mtu":1420,"address":["10.10.10.2/32"],"private_key":"aDHCHnkcdMjnq0bF+V4fARkbJBW8cWjuYoVjKfUwsXo=","peers":[{"address":"212.232.78.237","port":51820,"public_key":"fiK9ZG990zunr5cpRnx+SOVW2rVKKqFoVxmHMHAvAFk=","allowed_ips":["0.0.0.0/0","::/0"]}]}]}`)
	if err := os.WriteFile(p, body, 0644); err != nil {
		t.Fatal(err)
	}
	u, err := ShareProxyURIForOutboundTag(p, "wg1")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(u, "wireguard://") {
		t.Fatalf("expected wireguard URI, got %q", u)
	}
}

func TestGetEndpointMapByTag(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "config.json")
	if err := os.WriteFile(p, []byte(`{"endpoints":[{"type":"wireguard","tag":"e1","private_key":"x","address":["10.0.0.1/32"],"peers":[{"address":"1.1.1.1","port":51820,"public_key":"pub","allowed_ips":["0.0.0.0/0"]}]}]}`), 0644); err != nil {
		t.Fatal(err)
	}
	m, err := GetEndpointMapByTag(p, "e1")
	if err != nil {
		t.Fatal(err)
	}
	if mapGetStringTest(m, "tag") != "e1" {
		t.Fatalf("tag: %+v", m)
	}
}

func mapGetStringTest(m map[string]interface{}, k string) string {
	v, ok := m[k]
	if !ok {
		return ""
	}
	s, _ := v.(string)
	return s
}
