package locale

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

// placeholderRe matches Go format verbs like %s, %d, %v, %f, %q, %x, %02d, etc.
var placeholderRe = regexp.MustCompile(`%[-+# 0]*[*]?[0-9]*[.*]?[0-9]*[vTtbcdoOqxXUeEfFgGsp%]`)

func findProjectRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	dir := wd
	for i := 0; i < 10; i++ {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Fatalf("Project root not found from %s", wd)
	return ""
}

func loadExternalLocalesForTest(t *testing.T) {
	t.Helper()
	root := findProjectRoot(t)
	localeDir := filepath.Join(root, "bin", "locale")
	LoadExternalLocales(localeDir)
}

func TestEmbeddedEnglish(t *testing.T) {
	en, ok := catalogs["en"]
	if !ok {
		t.Fatal("embedded en catalog not found")
	}
	if len(en) < 10 {
		t.Errorf("en catalog too small: %d keys", len(en))
	}
	if name, ok := en[displayNameKey]; !ok || name != "English" {
		t.Errorf("en._display_name = %q, want %q", name, "English")
	}
}

func TestExternalRussian(t *testing.T) {
	loadExternalLocalesForTest(t)

	ru, ok := catalogs["ru"]
	if !ok {
		t.Skip("ru.json not found in bin/locale/ — skipping")
	}
	if name, ok := ru[displayNameKey]; !ok || name != "Русский" {
		t.Errorf("ru._display_name = %q, want %q", name, "Русский")
	}
}

func TestAllKeysPresent(t *testing.T) {
	loadExternalLocalesForTest(t)

	en := catalogs["en"]
	ru, ok := catalogs["ru"]
	if !ok {
		t.Skip("ru.json not found in bin/locale/ — skipping key completeness test")
	}

	for key := range en {
		if key == displayNameKey {
			continue
		}
		if _, ok := ru[key]; !ok {
			t.Errorf("key %q exists in en.json but missing in ru.json", key)
		}
	}
	for key := range ru {
		if key == displayNameKey {
			continue
		}
		if _, ok := en[key]; !ok {
			t.Errorf("key %q exists in ru.json but missing in en.json", key)
		}
	}
}

func TestNoEmptyValues(t *testing.T) {
	loadExternalLocalesForTest(t)

	for lang, msgs := range catalogs {
		for key, val := range msgs {
			if val == "" {
				t.Errorf("[%s] key %q has empty value", lang, key)
			}
		}
	}
}

func TestPlaceholderCount(t *testing.T) {
	loadExternalLocalesForTest(t)

	en := catalogs["en"]
	ru, ok := catalogs["ru"]
	if !ok {
		t.Skip("ru.json not found — skipping placeholder test")
	}

	for key, enVal := range en {
		if key == displayNameKey {
			continue
		}
		ruVal, ok := ru[key]
		if !ok {
			continue
		}
		enCount := countPlaceholders(enVal)
		ruCount := countPlaceholders(ruVal)
		if enCount != ruCount {
			t.Errorf("key %q: en has %d placeholder(s) (%q), ru has %d (%q)",
				key, enCount, enVal, ruCount, ruVal)
		}
	}
}

func TestTFunction(t *testing.T) {
	loadExternalLocalesForTest(t)

	SetLang("en")
	if got := T("core.button_start"); got != "Start" {
		t.Errorf("T(core.button_start) = %q, want %q", got, "Start")
	}

	if _, ok := catalogs["ru"]; ok {
		SetLang("ru")
		if got := T("core.button_start"); got != "Старт" {
			t.Errorf("T(core.button_start) = %q, want %q", got, "Старт")
		}
	}

	// Unknown key returns the key itself
	SetLang("en")
	if got := T("nonexistent.key"); got != "nonexistent.key" {
		t.Errorf("T(nonexistent.key) = %q, want %q", got, "nonexistent.key")
	}
}

func TestTfFunction(t *testing.T) {
	SetLang("en")
	got := Tf("help.version_label", "v1.0")
	want := fmt.Sprintf("📦 Version: %s", "v1.0")
	if got != want {
		t.Errorf("Tf(help.version_label, v1.0) = %q, want %q", got, want)
	}
}

func TestLanguages(t *testing.T) {
	langs := Languages()
	if len(langs) < 1 {
		t.Errorf("expected at least 1 language, got %d", len(langs))
	}
	found := false
	for _, l := range langs {
		if l == "en" {
			found = true
		}
	}
	if !found {
		t.Error("'en' not in Languages()")
	}
}

func TestLangDisplayName(t *testing.T) {
	if got := LangDisplayName("en"); got != "English" {
		t.Errorf("LangDisplayName(en) = %q, want %q", got, "English")
	}
}

func TestLangDisplayNameFromExternal(t *testing.T) {
	loadExternalLocalesForTest(t)
	if _, ok := catalogs["ru"]; !ok {
		t.Skip("ru.json not found")
	}
	if got := LangDisplayName("ru"); got != "Русский" {
		t.Errorf("LangDisplayName(ru) = %q, want %q", got, "Русский")
	}
}

func countPlaceholders(s string) int {
	matches := placeholderRe.FindAllString(s, -1)
	count := 0
	for _, m := range matches {
		if m != "%%" {
			count++
		}
	}
	return count
}
