package transcriber

import (
	"testing"
)

func TestNormalizeModelName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"models/gemini-2.5-flash", "gemini-2.5-flash"},
		{"gemini-2.5-flash", "gemini-2.5-flash"},
		{"gemini-3.0-flash", "gemini-3-flash-preview"}, // alias
		{"models/gemini-3.0-flash", "gemini-3-flash-preview"},
		{" gemini-2.5-flash ", "gemini-2.5-flash"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeModelName(tt.input)
			if got != tt.want {
				t.Errorf("normalizeModelName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestBuildModelCandidates(t *testing.T) {
	// Without env var, should return preferred models
	t.Setenv(ModelEnvVar, "")
	candidates := buildModelCandidates()
	if len(candidates) != len(PreferredModels) {
		t.Errorf("expected %d candidates, got %d", len(PreferredModels), len(candidates))
	}
	if candidates[0] != PreferredModels[0] {
		t.Errorf("first candidate should be %s, got %s", PreferredModels[0], candidates[0])
	}
}

func TestBuildModelCandidatesWithEnv(t *testing.T) {
	t.Setenv(ModelEnvVar, "my-custom-model")
	candidates := buildModelCandidates()
	if candidates[0] != "my-custom-model" {
		t.Errorf("first candidate should be env model, got %s", candidates[0])
	}
	if len(candidates) != len(PreferredModels)+1 {
		t.Errorf("expected %d candidates, got %d", len(PreferredModels)+1, len(candidates))
	}
}

func TestBuildModelCandidatesWithAliasEnv(t *testing.T) {
	t.Setenv(ModelEnvVar, "gemini-3.0-flash")
	candidates := buildModelCandidates()
	// Should be normalized to the alias
	if candidates[0] != "gemini-3-flash-preview" {
		t.Errorf("first candidate should be normalized alias, got %s", candidates[0])
	}
}
