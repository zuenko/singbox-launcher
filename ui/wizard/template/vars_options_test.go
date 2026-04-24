package template

import (
	"encoding/json"
	"testing"
)

func TestTemplateVarOptionsLegacyStringList(t *testing.T) {
	raw := `{"name":"log_level","type":"enum","options":["debug","info","warn"]}`
	var v TemplateVar
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got, want := v.Options, []string{"debug", "info", "warn"}; !sliceEq(got, want) {
		t.Errorf("Options = %v, want %v", got, want)
	}
	if v.OptionTitles != nil {
		t.Errorf("OptionTitles = %v, want nil (no titles when legacy form)", v.OptionTitles)
	}
	if v.OptionTitle(1) != "info" {
		t.Errorf("OptionTitle(1) = %q, want fallback to value %q", v.OptionTitle(1), "info")
	}
}

func TestTemplateVarOptionsObjectList(t *testing.T) {
	raw := `{"name":"urltest_interval","type":"text","options":[
		{"title":"5m (default)","value":"5m"},
		{"title":"30m (battery)","value":"30m"}
	]}`
	var v TemplateVar
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got, want := v.Options, []string{"5m", "30m"}; !sliceEq(got, want) {
		t.Errorf("Options = %v, want %v", got, want)
	}
	if got, want := v.OptionTitles, []string{"5m (default)", "30m (battery)"}; !sliceEq(got, want) {
		t.Errorf("OptionTitles = %v, want %v", got, want)
	}
	if v.OptionTitle(0) != "5m (default)" {
		t.Errorf("OptionTitle(0) = %q, want %q", v.OptionTitle(0), "5m (default)")
	}
}

func TestTemplateVarOptionsMixedList(t *testing.T) {
	// string among objects — each element is parsed independently.
	raw := `{"name":"mix","type":"text","options":["plain",{"title":"Fancy","value":"fancy"}]}`
	var v TemplateVar
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got, want := v.Options, []string{"plain", "fancy"}; !sliceEq(got, want) {
		t.Errorf("Options = %v, want %v", got, want)
	}
	if got, want := v.OptionTitles, []string{"plain", "Fancy"}; !sliceEq(got, want) {
		t.Errorf("OptionTitles = %v, want %v", got, want)
	}
}

func TestTemplateVarOptionsEmptyTitleFallsBackToValue(t *testing.T) {
	raw := `{"name":"x","type":"text","options":[{"title":"","value":"ok"}]}`
	var v TemplateVar
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if v.OptionTitle(0) != "ok" {
		t.Errorf("OptionTitle(0) = %q, want %q (title='' falls back to value)", v.OptionTitle(0), "ok")
	}
}

func sliceEq(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
