package locale

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"singbox-launcher/internal/debuglog"
)

// Settings represents the launcher settings stored in bin/settings.json.
type Settings struct {
	Lang string `json:"lang"`
}

// LoadSettings reads settings from binDir/settings.json.
// Returns default settings if file doesn't exist or is invalid.
func LoadSettings(binDir string) Settings {
	s := Settings{Lang: "en"}
	path := filepath.Join(binDir, "settings.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return s
	}
	if err := json.Unmarshal(data, &s); err != nil {
		debuglog.WarnLog("locale: failed to parse settings.json: %v", err)
		return Settings{Lang: "en"}
	}
	if s.Lang == "" {
		s.Lang = "en"
	}
	return s
}

// SaveSettings writes settings to binDir/settings.json.
func SaveSettings(binDir string, s Settings) error {
	path := filepath.Join(binDir, "settings.json")
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("locale: marshal settings: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}
