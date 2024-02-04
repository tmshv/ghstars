package icons

import (
	"testing"
)

func TestLang(t *testing.T) {
	nerd := Nerd()
	tests := []struct {
		lang     string
		expected string
	}{
		{"Python", "\ue73c"},
		{"TypeScript", "\ue628"},
		{"Rust", "\ue7a8"},
		{"Go", "\ue65e"},
		{"C", "\ue649"},
		{"Unknown", "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.lang, func(t *testing.T) {
			result := nerd.Lang(tt.lang)
			if result != tt.expected {
				t.Errorf("Lang(%q) = %q, want %q", tt.lang, result, tt.expected)
			}
		})
	}
}
