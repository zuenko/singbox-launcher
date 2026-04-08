package config

import (
	"encoding/json"
	"testing"
)

func TestExperimentalCacheFileFromSection(t *testing.T) {
	raw := json.RawMessage(`{"cache_file":{"enabled":true,"path":"cache.db"}}`)
	ok, p := ExperimentalCacheFileFromSection(raw)
	if !ok || p != "cache.db" {
		t.Fatalf("got ok=%v p=%q", ok, p)
	}
	disabled := json.RawMessage(`{"cache_file":{"enabled":false,"path":"cache.db"}}`)
	ok2, _ := ExperimentalCacheFileFromSection(disabled)
	if ok2 {
		t.Fatal("disabled cache_file should not remove")
	}
}
