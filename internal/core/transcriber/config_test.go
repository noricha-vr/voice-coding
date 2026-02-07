package transcriber

import (
	"testing"
	"time"
)

func TestResolvePromptCacheTTL(t *testing.T) {
	tests := []struct {
		name string
		env  string
		want time.Duration
	}{
		{"empty uses default", "", DefaultPromptCacheTTL},
		{"valid duration", "30m", 30 * time.Minute},
		{"valid seconds", "120s", 120 * time.Second},
		{"invalid falls back to default", "notaduration", DefaultPromptCacheTTL},
		{"bare number falls back to default", "3600", DefaultPromptCacheTTL},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(PromptCacheTTLEnvVar, tt.env)
			got := resolvePromptCacheTTL()
			if got != tt.want {
				t.Errorf("resolvePromptCacheTTL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResolveThinkingLevel(t *testing.T) {
	tests := []struct {
		name string
		env  string
		want string // compare by mapped key name
	}{
		{"empty uses default", "", DefaultThinkingLevel},
		{"minimal", "minimal", "minimal"},
		{"high", "high", "high"},
		{"invalid falls back to default", "invalid_level", DefaultThinkingLevel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(ThinkingLevelEnvVar, tt.env)
			got := resolveThinkingLevel()
			want := thinkingLevelMap[tt.want]
			if got != want {
				t.Errorf("resolveThinkingLevel() = %v, want %v", got, want)
			}
		})
	}
}

func TestResolvePromptCacheEnabled(t *testing.T) {
	tests := []struct {
		name string
		env  string
		want bool
	}{
		{"empty defaults to true", "", true},
		{"false", "false", false},
		{"0", "0", false},
		{"off", "off", false},
		{"no", "no", false},
		{"true", "true", true},
		{"1", "1", true},
		{"random string", "yes", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(EnablePromptCacheEnvVar, tt.env)
			got := resolvePromptCacheEnabled()
			if got != tt.want {
				t.Errorf("resolvePromptCacheEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}
