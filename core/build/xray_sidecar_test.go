package build

import (
	"encoding/json"
	"strings"
	"testing"

	"singbox-launcher/core/template"
	"singbox-launcher/core/xray"
)

func TestApplyXraySidecarTransform_SidecarPath(t *testing.T) {
	outbounds := []json.RawMessage{
		[]byte(`{"tag":"tcp-node","type":"vless","server":"a.com","server_port":443,"uuid":"u1"}`),
		[]byte(`{"tag":"xh-node","type":"vless","server":"b.com","server_port":443,"uuid":"u2","transport":{"type":"xhttp","path":"/p","host":"h","mode":"auto"}}`),
	}
	cache := &ParsedCache{Outbounds: outbounds}
	ctx := BuildContext{
		Cache: cache,
		XraySidecar: XraySidecarConfig{
			Enabled:  true,
			XrayPath: "/nonexistent/xray", // fallback path
			BasePort: 15080,
		},
	}

	registry := applyXraySidecarTransform(&ctx)
	if registry != nil {
		// xray binary doesn't exist, so registry should be nil (fallback applied)
		t.Fatalf("expected nil registry when xray missing, got %+v", registry)
	}

	// Verify fallback: xhttp → httpupgrade, mode/extra removed
	var ob map[string]interface{}
	if err := json.Unmarshal(ctx.Cache.Outbounds[1], &ob); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	tr := ob["transport"].(map[string]interface{})
	if tr["type"] != "httpupgrade" {
		t.Fatalf("expected httpupgrade fallback, got %v", tr["type"])
	}
	if _, ok := tr["mode"]; ok {
		t.Fatalf("mode should be removed in fallback")
	}
}

func TestApplyXraySidecarTransform_Disabled(t *testing.T) {
	outbounds := []json.RawMessage{
		[]byte(`{"tag":"xh-node","type":"vless","server":"b.com","server_port":443,"uuid":"u2","transport":{"type":"xhttp","path":"/p"}}`),
	}
	cache := &ParsedCache{Outbounds: outbounds}
	ctx := BuildContext{
		Cache: cache,
		XraySidecar: XraySidecarConfig{
			Enabled: false,
		},
	}

	registry := applyXraySidecarTransform(&ctx)
	if registry != nil {
		t.Fatalf("expected nil registry when disabled, got %+v", registry)
	}

	var ob map[string]interface{}
	if err := json.Unmarshal(ctx.Cache.Outbounds[0], &ob); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	tr := ob["transport"].(map[string]interface{})
	if tr["type"] != "httpupgrade" {
		t.Fatalf("expected httpupgrade fallback when disabled, got %v", tr["type"])
	}
}

func TestApplyXraySidecarTransform_RegistryAndSocks(t *testing.T) {
	// This test simulates the sidecar path by using the current executable as xray binary
	// (it won't actually run, but fileExists will pass).
	outbounds := []json.RawMessage{
		[]byte(`{"tag":"tcp-node","type":"vless","server":"a.com","server_port":443,"uuid":"u1"}`),
		[]byte(`{"tag":"xh-node","type":"vless","server":"b.com","server_port":443,"uuid":"u2","transport":{"type":"xhttp","path":"/p","host":"h","mode":"auto"}}`),
	}
	cache := &ParsedCache{Outbounds: outbounds}

	// Use a path that exists (the test binary itself) so fileExists passes
	ctx := BuildContext{
		Cache: cache,
		XraySidecar: XraySidecarConfig{
			Enabled:  true,
			XrayPath: "/bin/sh", // exists on unix; on windows this won't exist
			BasePort: 15080,
		},
	}

	// Skip on Windows since /bin/sh doesn't exist
	if ctx.XraySidecar.XrayPath == "/bin/sh" {
		if !fileExists("/bin/sh") {
			t.Skip("skipping: /bin/sh not available")
		}
	}

	registry := applyXraySidecarTransform(&ctx)
	if registry == nil {
		t.Fatalf("expected registry when xray exists")
	}
	if registry.Len() != 1 {
		t.Fatalf("expected 1 registry entry, got %d", registry.Len())
	}

	entry, ok := registry.Get("xh-node")
	if !ok {
		t.Fatalf("expected xh-node in registry")
	}
	if entry.Port != 15080 {
		t.Fatalf("expected port 15080, got %d", entry.Port)
	}

	// Verify the outbound was replaced with socks
	var ob map[string]interface{}
	if err := json.Unmarshal(ctx.Cache.Outbounds[1], &ob); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if ob["type"] != "socks" {
		t.Fatalf("expected socks type, got %v", ob["type"])
	}
	if ob["server"] != "127.0.0.1" {
		t.Fatalf("expected server 127.0.0.1, got %v", ob["server"])
	}
	if int(ob["server_port"].(float64)) != 15080 {
		t.Fatalf("expected port 15080, got %v", ob["server_port"])
	}

	// Verify original outbound preserved in registry
	orig := entry.OriginalOutbound
	if orig["type"] != "vless" {
		t.Fatalf("original outbound type mismatch")
	}
}

func TestBuildConfig_WithSidecarRegistry(t *testing.T) {
	td := &template.TemplateData{
		Config: map[string]json.RawMessage{
			"outbounds": []byte(`[]`),
		},
		ConfigOrder: []string{"outbounds"},
	}
	outbounds := []json.RawMessage{
		[]byte(`{"tag":"xh-node","type":"vless","server":"b.com","server_port":443,"uuid":"u2","transport":{"type":"xhttp","path":"/p"}}`),
	}
	ctx := BuildContext{
		Template:   td,
		Cache:      &ParsedCache{Outbounds: outbounds},
		ForPreview: false,
		XraySidecar: XraySidecarConfig{
			Enabled: false, // fallback path
		},
	}

	res, err := BuildConfig(ctx)
	if err != nil {
		t.Fatalf("BuildConfig: %v", err)
	}

	// With sidecar disabled, registry should be nil (fallback applied instead)
	if res.SidecarRegistry != nil {
		t.Fatalf("expected nil registry when sidecar disabled")
	}

	// Check that the final JSON contains httpupgrade (fallback applied)
	if !strings.Contains(string(res.ConfigJSON), `"type":"httpupgrade"`) {
		t.Fatalf("expected httpupgrade in final config JSON: %s", string(res.ConfigJSON))
	}
}

func TestNewRegistry(t *testing.T) {
	r := xray.NewRegistry(15080)
	if r.Len() != 0 {
		t.Fatalf("expected empty registry")
	}
}
