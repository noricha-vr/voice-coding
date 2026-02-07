package transcriber

import (
	"context"
	"log"
	"os"
	"strings"

	"google.golang.org/genai"
)

const (
	DefaultModel = "gemini-2.5-flash"
	ModelEnvVar  = "VOICECODE_GEMINI_MODEL"
)

// PreferredModels lists models in priority order.
var PreferredModels = []string{
	"gemini-3-flash-preview",
	DefaultModel,
	"gemini-2.0-flash",
	"gemini-flash-latest",
}

// ModelAliases maps old model names to current names.
var ModelAliases = map[string]string{
	"gemini-3.0-flash": "gemini-3-flash-preview",
}

// normalizeModelName removes "models/" prefix and applies aliases.
func normalizeModelName(name string) string {
	name = strings.TrimPrefix(name, "models/")
	name = strings.TrimSpace(name)
	if alias, ok := ModelAliases[name]; ok {
		return alias
	}
	return name
}

// buildModelCandidates returns model candidates in priority order.
func buildModelCandidates() []string {
	configured := strings.TrimSpace(os.Getenv(ModelEnvVar))
	var candidates []string

	if configured != "" {
		candidates = append(candidates, normalizeModelName(configured))
	}

	for _, m := range PreferredModels {
		found := false
		for _, c := range candidates {
			if c == m {
				found = true
				break
			}
		}
		if !found {
			candidates = append(candidates, m)
		}
	}

	return candidates
}

// listAvailableModels fetches models that support generateContent from the API.
func listAvailableModels(ctx context.Context, client *genai.Client) ([]string, error) {
	page, err := client.Models.List(ctx, nil)
	if err != nil {
		return nil, err
	}

	var available []string
	for _, model := range page.Items {
		for _, action := range model.SupportedActions {
			if action == "generateContent" {
				available = append(available, normalizeModelName(model.Name))
				break
			}
		}
	}
	return available, nil
}

// ResolveModel determines the best model to use.
func ResolveModel(ctx context.Context, client *genai.Client, exclude map[string]bool) string {
	if exclude == nil {
		exclude = make(map[string]bool)
	}
	candidates := buildModelCandidates()
	configured := strings.TrimSpace(os.Getenv(ModelEnvVar))
	if configured != "" {
		configured = normalizeModelName(configured)
	}

	available, err := listAvailableModels(ctx, client)
	if err != nil {
		if configured != "" && !exclude[configured] {
			log.Printf("[Gemini] モデル一覧取得に失敗したため環境変数モデルを使用します: %s (%v)", configured, err)
			return configured
		}
		if !exclude[DefaultModel] {
			log.Printf("[Gemini] モデル一覧取得に失敗したため既定モデルを使用します: %s (%v)", DefaultModel, err)
			return DefaultModel
		}
		return ""
	}

	availSet := make(map[string]bool, len(available))
	for _, m := range available {
		availSet[m] = true
	}

	for _, c := range candidates {
		if availSet[c] && !exclude[c] {
			return c
		}
	}

	for _, m := range available {
		if !exclude[m] {
			log.Printf("[Gemini] 候補モデルが見つからないため利用可能モデルにフォールバックします: %s", m)
			return m
		}
	}

	return ""
}
