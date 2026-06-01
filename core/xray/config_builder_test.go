package xray

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestBuildSidecarConfig_Basic(t *testing.T) {
	sb := map[string]interface{}{
		"type":        "vless",
		"server":      "example.com",
		"server_port": 443,
		"uuid":        "550e8400-e29b-41d4-a716-446655440000",
		"transport": map[string]interface{}{
			"type": "xhttp",
			"path": "/path",
			"host": "cdn.example",
			"mode": "auto",
		},
	}

	cfgJSON, err := BuildSidecarConfig(sb, 15080)
	if err != nil {
		t.Fatalf("BuildSidecarConfig: %v", err)
	}

	var cfg map[string]interface{}
	if err := json.Unmarshal(cfgJSON, &cfg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	// Check inbound
	inbounds := cfg["inbounds"].([]interface{})
	if len(inbounds) != 1 {
		t.Fatalf("expected 1 inbound, got %d", len(inbounds))
	}
	inb := inbounds[0].(map[string]interface{})
	if inb["protocol"] != "socks" || inb["port"] != float64(15080) {
		t.Fatalf("inbound mismatch: %+v", inb)
	}

	// Check outbound
	outbounds := cfg["outbounds"].([]interface{})
	if len(outbounds) != 2 {
		t.Fatalf("expected 2 outbounds, got %d", len(outbounds))
	}
	out := outbounds[0].(map[string]interface{})
	if out["protocol"] != "vless" {
		t.Fatalf("expected vless outbound, got %v", out["protocol"])
	}

	ss := out["streamSettings"].(map[string]interface{})
	if ss["network"] != "xhttp" {
		t.Fatalf("expected xhttp network, got %v", ss["network"])
	}
	xh := ss["xhttpSettings"].(map[string]interface{})
	if xh["path"] != "/path" || xh["host"] != "cdn.example" || xh["mode"] != "auto" {
		t.Fatalf("xhttpSettings mismatch: %+v", xh)
	}
}

func TestBuildSidecarConfig_Reality(t *testing.T) {
	sb := map[string]interface{}{
		"type":        "vless",
		"server":      "srv.example",
		"server_port": 443,
		"uuid":        "550e8400-e29b-41d4-a716-446655440000",
		"flow":        "xtls-rprx-vision",
		"transport": map[string]interface{}{
			"type": "xhttp",
			"path": "/hx",
		},
		"tls": map[string]interface{}{
			"enabled":     true,
			"server_name": "sni.example",
			"reality": map[string]interface{}{
				"enabled":     true,
				"public_key":  "AbCdEf123",
				"short_id":    "abc123",
				"fingerprint": "chrome",
			},
		},
	}

	cfgJSON, err := BuildSidecarConfig(sb, 15081)
	if err != nil {
		t.Fatalf("BuildSidecarConfig: %v", err)
	}

	var cfg map[string]interface{}
	if err := json.Unmarshal(cfgJSON, &cfg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	outbounds := cfg["outbounds"].([]interface{})
	out := outbounds[0].(map[string]interface{})
	ss := out["streamSettings"].(map[string]interface{})

	if ss["security"] != "reality" {
		t.Fatalf("expected reality security, got %v", ss["security"])
	}
	rs := ss["realitySettings"].(map[string]interface{})
	if rs["publicKey"] != "AbCdEf123" || rs["shortId"] != "abc123" {
		t.Fatalf("realitySettings mismatch: %+v", rs)
	}
}

func TestBuildSidecarConfig_TLS(t *testing.T) {
	sb := map[string]interface{}{
		"type":        "vless",
		"server":      "srv.example",
		"server_port": 443,
		"uuid":        "550e8400-e29b-41d4-a716-446655440000",
		"transport": map[string]interface{}{
			"type": "xhttp",
			"path": "/hx",
		},
		"tls": map[string]interface{}{
			"enabled":     true,
			"server_name": "sni.example",
			"insecure":    true,
		},
	}

	cfgJSON, err := BuildSidecarConfig(sb, 15082)
	if err != nil {
		t.Fatalf("BuildSidecarConfig: %v", err)
	}

	var cfg map[string]interface{}
	if err := json.Unmarshal(cfgJSON, &cfg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	outbounds := cfg["outbounds"].([]interface{})
	out := outbounds[0].(map[string]interface{})
	ss := out["streamSettings"].(map[string]interface{})

	if ss["security"] != "tls" {
		t.Fatalf("expected tls security, got %v", ss["security"])
	}
	ts := ss["tlsSettings"].(map[string]interface{})
	if ts["allowInsecure"] != true {
		t.Fatalf("expected allowInsecure=true, got %v", ts["allowInsecure"])
	}
}

func TestBuildSidecarConfig_UnsupportedType(t *testing.T) {
	sb := map[string]interface{}{
		"type":        "trojan",
		"server":      "example.com",
		"server_port": 443,
	}
	_, err := BuildSidecarConfig(sb, 15080)
	if err == nil || !strings.Contains(err.Error(), "unsupported outbound type") {
		t.Fatalf("expected unsupported type error, got %v", err)
	}
}

func TestRegistry(t *testing.T) {
	r := NewRegistry(15080)
	port, socks := r.Register("node-xh", map[string]interface{}{"type": "vless"})
	if port != 15080 {
		t.Fatalf("expected port 15080, got %d", port)
	}
	if socks["type"] != "socks" {
		t.Fatalf("expected socks outbound, got %v", socks["type"])
	}

	port2, _ := r.Register("node-xh2", map[string]interface{}{"type": "vless"})
	if port2 != 15081 {
		t.Fatalf("expected port 15081, got %d", port2)
	}

	// Re-register same tag → same port
	port3, _ := r.Register("node-xh", map[string]interface{}{"type": "vless"})
	if port3 != 15080 {
		t.Fatalf("expected re-used port 15080, got %d", port3)
	}

	if r.Len() != 2 {
		t.Fatalf("expected len 2, got %d", r.Len())
	}
}
