// Package urlredact removes user credentials embedded in URL-like substrings (for safe error messages).
package urlredact

import "regexp"

// credentialsInURLPattern matches scheme://user:password@ in common URL forms.
var credentialsInURLPattern = regexp.MustCompile(`([a-zA-Z][a-zA-Z0-9+\-.]*://)([^:@/\s]+):([^@/\s]+)@`)

// RedactURLUserinfo replaces password segments in matched URLs with ***.
func RedactURLUserinfo(message string) string {
	return credentialsInURLPattern.ReplaceAllString(message, "${1}${2}:***@")
}

// RedactToken masks an opaque secret (API bearer token, Clash API `secret`, etc.)
// for safe inclusion in log output. Short tokens (≤ 6 chars) are fully masked;
// longer tokens show the first 2 and last 2 characters with the middle elided.
// Examples:
//
//	"abc"                  → "***"
//	"abcdef12"             → "ab***12"
//	"sk-live-0123456789"   → "sk***89"
func RedactToken(token string) string {
	if len(token) <= 6 {
		return "***"
	}
	return token[:2] + "***" + token[len(token)-2:]
}
