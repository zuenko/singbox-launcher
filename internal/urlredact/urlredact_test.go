package urlredact

import "testing"

func TestRedactURLUserinfo(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "http userinfo",
			in:   `dial tcp: connect failed: http://user:secret@proxy.example:8080`,
			want: `dial tcp: connect failed: http://user:***@proxy.example:8080`,
		},
		{
			name: "https userinfo",
			in:   `Get "https://admin:pass@cdn.example/path": timeout`,
			want: `Get "https://admin:***@cdn.example/path": timeout`,
		},
		{
			name: "no change without userinfo",
			in:   `Network error: connection refused`,
			want: `Network error: connection refused`,
		},
		{
			name: "multiple occurrences",
			in:   `a: http://u1:p1@h1 b: https://u2:p2@h2`,
			want: `a: http://u1:***@h1 b: https://u2:***@h2`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RedactURLUserinfo(tt.in)
			if got != tt.want {
				t.Fatalf("RedactURLUserinfo(%q) = %q; want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestRedactToken(t *testing.T) {
	tests := []struct {
		name  string
		token string
		want  string
	}{
		{"empty", "", "***"},
		{"very short", "abc", "***"},
		{"exactly 6 chars", "abcdef", "***"},
		{"7 chars", "abcdef1", "ab***f1"},
		{"typical 32-char secret", "s3cr3t-t0k3n-0123456789abcdef123", "s3***23"},
		{"prefixed API key", "sk-live-0123456789", "sk***89"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RedactToken(tt.token); got != tt.want {
				t.Errorf("RedactToken(%q) = %q; want %q", tt.token, got, tt.want)
			}
		})
	}
}
