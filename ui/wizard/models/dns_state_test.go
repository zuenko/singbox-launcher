package models

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestPersistedDNSState_JSONStrategyRoundTrip(t *testing.T) {
	in := &PersistedDNSState{
		Servers: []json.RawMessage{json.RawMessage(`{"tag":"a","type":"udp","server":"1.1.1.1"}`)},
		Rules: []json.RawMessage{
			json.RawMessage(`{"server":"a"}`),
		},
		Final:    "a",
		Strategy: "prefer_ipv4",
	}
	data, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	if !strings.Contains(s, `"strategy"`) || !strings.Contains(s, `"prefer_ipv4"`) {
		t.Fatalf("expected dns_options.strategy in JSON, got: %s", s)
	}
	if !strings.Contains(s, `"rules"`) {
		t.Fatalf("expected dns_options.rules in JSON, got: %s", s)
	}
	var out PersistedDNSState
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatal(err)
	}
	if out.Strategy != "prefer_ipv4" {
		t.Fatalf("Strategy after unmarshal: got %q want prefer_ipv4", out.Strategy)
	}
	if len(out.Rules) != 1 || string(out.Rules[0]) != `{"server":"a"}` {
		t.Fatalf("Rules after unmarshal: %+v", out.Rules)
	}
}

func TestPersistedDNSState_StrategyOmitemptyWhenEmpty(t *testing.T) {
	in := &PersistedDNSState{
		Servers: []json.RawMessage{json.RawMessage(`{"tag":"a","type":"udp","server":"1.1.1.1"}`)},
		Final:   "a",
	}
	data, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), `"strategy"`) {
		t.Fatalf("empty strategy should omit json key, got: %s", data)
	}
}
