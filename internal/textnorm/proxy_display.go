package textnorm

import "strings"

// NormalizeProxyDisplay repairs invalid UTF-8 and replaces Unicode angle quotation marks
// common in subscription/outbound tags (U+276F ❯, U+00BB », U+203A ›) with ASCII " > ".
// Collapses runs of spaces and trims ends. Safe for tags sent to sing-box and for UI labels.
func NormalizeProxyDisplay(s string) string {
	if s == "" {
		return s
	}
	s = strings.ToValidUTF8(s, "")
	s = strings.ReplaceAll(s, "\u276f", " > ") // ❯ HEAVY RIGHT-POINTING ANGLE QUOTATION MARK
	s = strings.ReplaceAll(s, "\u00bb", " > ") // » RIGHT-POINTING DOUBLE ANGLE QUOTATION MARK
	s = strings.ReplaceAll(s, "\u203a", " > ") // › SINGLE RIGHT-POINTING ANGLE QUOTATION MARK
	for strings.Contains(s, "  ") {
		s = strings.ReplaceAll(s, "  ", " ")
	}
	return strings.TrimSpace(s)
}
