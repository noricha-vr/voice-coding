package transcriber

import (
	"errors"
	"testing"
)

func TestIsModelNotFound(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"api version error", errors.New("models/gemini-x is not found for api version v1beta"), true},
		{"404 not found", errors.New("404 models/gemini-x not found"), true},
		{"unrelated error", errors.New("connection refused"), false},
		{"partial match", errors.New("404 something else"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsModelNotFound(tt.err); got != tt.want {
				t.Errorf("IsModelNotFound(%q) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

func TestIsTransient(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"timeout", errors.New("request timed out"), true},
		{"deadline", errors.New("deadline expired"), true},
		{"429", errors.New("HTTP 429 Too Many Requests"), true},
		{"503", errors.New("503 service unavailable"), true},
		{"too many requests", errors.New("too many requests"), true},
		{"normal error", errors.New("invalid argument"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsTransient(tt.err); got != tt.want {
				t.Errorf("IsTransient(%q) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

func TestIsThinkingUnsupported(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"unsupported", errors.New("thinking level is not supported for this model"), true},
		{"other", errors.New("some other error"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsThinkingUnsupported(tt.err); got != tt.want {
				t.Errorf("IsThinkingUnsupported(%q) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

func TestIsCachedContentError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"not found", errors.New("CachedContent not found"), true},
		{"permission", errors.New("PERMISSION_DENIED: cached content access denied"), true},
		{"other", errors.New("invalid request"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsCachedContentError(tt.err); got != tt.want {
				t.Errorf("IsCachedContentError(%q) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}
