package business

import (
	"encoding/json"
	"strings"
	"testing"

	"singbox-launcher/core/config"
)

// TestSerializeParserConfig_Standalone tests SerializeParserConfig without UI dependencies
func TestSerializeParserConfig_Standalone(t *testing.T) {
	tests := []struct {
		name        string
		config      *config.ParserConfig
		expectError bool
		checkResult func(*testing.T, string)
	}{
		{
			name: "Valid ParserConfig",
			config: &config.ParserConfig{
				ParserConfig: struct {
					Version   int                     `json:"version,omitempty"`
					Proxies   []config.ProxySource    `json:"proxies"`
					Outbounds []config.OutboundConfig `json:"outbounds"`
					Parser    struct {
						Reload      string `json:"reload,omitempty"`
						LastUpdated string `json:"last_updated,omitempty"`
					} `json:"parser,omitempty"`
				}{
					Version: 2,
					Proxies: []config.ProxySource{
						{
							Source:      "https://example.com/subscription",
							Connections: []string{"vless://uuid@server:443"},
						},
					},
					Outbounds: []config.OutboundConfig{
						{
							Tag:  "proxy-out",
							Type: "selector",
						},
					},
				},
			},
			expectError: false,
			checkResult: func(t *testing.T, result string) {
				if result == "" {
					t.Error("Expected non-empty result")
				}
				var parsed map[string]interface{}
				if err := json.Unmarshal([]byte(result), &parsed); err != nil {
					t.Errorf("Result is not valid JSON: %v", err)
				}
				if _, ok := parsed["ParserConfig"]; !ok {
					t.Error("Expected ParserConfig in result")
				}
			},
		},
		{
			name:        "Nil ParserConfig",
			config:      nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SerializeParserConfig(tt.config)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if tt.checkResult != nil {
				tt.checkResult(t, result)
			}
		})
	}
}

// TestApplyURLToParserConfig_Logic tests the logic of ApplyURLToParserConfig without UI
func TestApplyURLToParserConfig_Logic(t *testing.T) {
	// Test URL classification logic
	input := `https://example.com/subscription
vless://uuid@server:443#Test
https://another.com/sub
vmess://base64`

	lines := strings.Split(input, "\n")
	subscriptions := make([]string, 0)
	connections := make([]string, 0)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "http://") || strings.HasPrefix(line, "https://") {
			subscriptions = append(subscriptions, line)
		} else if strings.Contains(line, "://") {
			connections = append(connections, line)
		}
	}

	if len(subscriptions) != 2 {
		t.Errorf("Expected 2 subscriptions, got %d", len(subscriptions))
	}
	if len(connections) != 2 {
		t.Errorf("Expected 2 connections, got %d", len(connections))
	}
}

func TestTagPrefixFromSubscriptionFragment(t *testing.T) {
	tests := []struct {
		raw  string
		want string
	}{
		{"https://xray.example/v/126.json#abvpn", "abvpn:"},
		{"https://xray.example/v/126.json#my%20label", "my label:"},
		{"https://xray.example/v/126.json", ""},
		{"https://xray.example/v/126.json#", ""},
		{"https://xray.example/v/126.json#%09trim%09", "trim:"},
		{"vless://uuid@host:443#nope", ""},
	}
	for _, tt := range tests {
		t.Run(tt.raw, func(t *testing.T) {
			got := tagPrefixFromSubscriptionFragment(tt.raw)
			if got != tt.want {
				t.Errorf("tagPrefixFromSubscriptionFragment(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}

func TestBuildProxiesFromInputs_TagPrefixFromURLFragment(t *testing.T) {
	empty := &existingProperties{
		OutboundsMap:  make(map[string][]config.OutboundConfig),
		TagPrefixMap:  make(map[string]string),
		TagPostfixMap: make(map[string]string),
	}
	sub := "https://host/sub.json#abvpn"
	items := []proxyInput{{Subscription: sub}}
	got := buildProxiesFromInputs(items, empty, nil, 1)
	if len(got) != 1 || got[0].TagPrefix != "abvpn:" {
		t.Fatalf("got %+v", got)
	}
	// Saved tag_prefix wins over fragment
	withMap := &existingProperties{
		OutboundsMap:  make(map[string][]config.OutboundConfig),
		TagPrefixMap:  map[string]string{sub: "saved:"},
		TagPostfixMap: make(map[string]string),
	}
	got2 := buildProxiesFromInputs(items, withMap, nil, 1)
	if len(got2) != 1 || got2[0].TagPrefix != "saved:" {
		t.Fatalf("expected saved prefix, got %+v", got2)
	}
}
