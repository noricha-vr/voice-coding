package transcriber

import (
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/noricha-vr/voicecode/internal/core/prompt"
	"google.golang.org/genai"
)

const (
	ThinkingLevelEnvVar     = "VOICECODE_THINKING_LEVEL"
	EnablePromptCacheEnvVar = "VOICECODE_ENABLE_PROMPT_CACHE"
	PromptCacheTTLEnvVar    = "VOICECODE_PROMPT_CACHE_TTL"
	DefaultThinkingLevel    = "minimal"
	DefaultPromptCacheTTL   = 3600 * time.Second
	Timeout                 = 10 * time.Second
	MaxTransientRetries     = 1
	RetryBackoffSeconds     = 0.3
)

var xmlTagPattern = regexp.MustCompile(`<[^>]+>`)

var thinkingLevelMap = map[string]genai.ThinkingLevel{
	"minimal": genai.ThinkingLevelMinimal,
	"low":     genai.ThinkingLevelLow,
	"medium":  genai.ThinkingLevelMedium,
	"high":    genai.ThinkingLevelHigh,
}

// Transcriber handles audio transcription via Gemini API.
type Transcriber struct {
	client            *genai.Client
	systemPrompt      string
	modelName         string
	thinkingLevel     genai.ThinkingLevel
	thinkingMode      string // "level" or "budget0"
	enablePromptCache bool
	promptCacheTTL    time.Duration
	cacheNameByModel  map[string]string
}

// New creates and initializes a Transcriber.
func New(ctx context.Context, apiKey string) (*Transcriber, error) {
	if apiKey == "" {
		apiKey = os.Getenv("GOOGLE_API_KEY")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("GOOGLE_API_KEY is not set")
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("create genai client: %w", err)
	}

	t := &Transcriber{
		client:            client,
		systemPrompt:      prompt.SystemPrompt,
		thinkingLevel:     resolveThinkingLevel(),
		thinkingMode:      "level",
		enablePromptCache: resolvePromptCacheEnabled(),
		promptCacheTTL:    resolvePromptCacheTTL(),
		cacheNameByModel:  make(map[string]string),
	}

	t.modelName = ResolveModel(ctx, client, nil)
	if t.modelName == "" {
		return nil, fmt.Errorf("no available model found")
	}

	t.ensurePromptCache(ctx)
	log.Printf("[Gemini] 使用モデル: %s", t.modelName)
	log.Printf("[Gemini] Thinking mode: %s (%s)", t.thinkingMode, t.thinkingLevel)

	return t, nil
}

// Transcribe transcribes a WAV file and returns the text.
func (t *Transcriber) Transcribe(ctx context.Context, wavPath string) (string, float64, error) {
	audioData, err := os.ReadFile(wavPath)
	if err != nil {
		return "", 0, fmt.Errorf("read audio file: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, Timeout)
	defer cancel()

	start := time.Now()

	response, err := t.generateContentWithRetry(ctx, audioData)
	if err != nil {
		if IsModelNotFound(err) || IsTransient(err) {
			fallback := ResolveModel(ctx, t.client, map[string]bool{t.modelName: true})
			if fallback != "" {
				reason := "モデルが見つからないため"
				if IsTransient(err) {
					reason = "一時的なAPIエラーのため"
				}
				log.Printf("[Gemini] %sモデルを切替します: %s -> %s", reason, t.modelName, fallback)
				t.modelName = fallback
				t.ensurePromptCache(ctx)
				response, err = t.generateContentWithRetry(ctx, audioData)
				if err != nil {
					elapsed := time.Since(start).Seconds()
					log.Printf("[Gemini %.2fs] API呼び出しに失敗しました(model=%s): %v", elapsed, t.modelName, err)
					return "", elapsed, err
				}
			} else {
				elapsed := time.Since(start).Seconds()
				log.Printf("[Gemini %.2fs] API呼び出しに失敗しました(model=%s): %v", elapsed, t.modelName, err)
				return "", elapsed, err
			}
		} else {
			elapsed := time.Since(start).Seconds()
			log.Printf("[Gemini %.2fs] API呼び出しに失敗しました(model=%s): %v", elapsed, t.modelName, err)
			return "", elapsed, err
		}
	}

	elapsed := time.Since(start).Seconds()
	rawText := response.Text()
	result := xmlTagPattern.ReplaceAllString(rawText, "")
	result = strings.TrimSpace(result)

	log.Printf("[Gemini %.2fs] %s (model=%s)", elapsed, result, t.modelName)
	return result, elapsed, nil
}

// ModelName returns the current model name.
func (t *Transcriber) ModelName() string {
	return t.modelName
}

func (t *Transcriber) generateContentWithRetry(ctx context.Context, audioData []byte) (*genai.GenerateContentResponse, error) {
	resp, err := t.retryLoop(ctx, audioData)
	if err != nil && IsCachedContentError(err) {
		log.Printf("[Gemini] CachedContent が無効なためキャッシュなしで再試行します (model=%s)", t.modelName)
		delete(t.cacheNameByModel, t.modelName)
		return t.retryLoop(ctx, audioData)
	}
	return resp, err
}

func (t *Transcriber) retryLoop(ctx context.Context, audioData []byte) (*genai.GenerateContentResponse, error) {
	var lastErr error
	for attempt := 0; attempt <= MaxTransientRetries; attempt++ {
		resp, err := t.generateContent(ctx, audioData)
		if err == nil {
			return resp, nil
		}
		lastErr = err

		if t.thinkingMode == "level" && IsThinkingUnsupported(err) {
			t.thinkingMode = "budget0"
			log.Printf("[Gemini] thinking_level 非対応モデルのため thinking_budget=0 に切替します")
			return t.generateContent(ctx, audioData)
		}

		if !IsTransient(err) || attempt >= MaxTransientRetries {
			return nil, err
		}

		wait := time.Duration(float64(time.Second) * RetryBackoffSeconds * float64(attempt+1))
		log.Printf("[Gemini] 一時的なAPIエラーのため再試行します(%d/%d): %v", attempt+1, MaxTransientRetries, err)
		time.Sleep(wait)
	}
	return nil, lastErr
}

func (t *Transcriber) generateContent(ctx context.Context, audioData []byte) (*genai.GenerateContentResponse, error) {
	contents := []*genai.Content{
		{
			Parts: []*genai.Part{
				{Text: prompt.TranscribePrompt},
				{InlineData: &genai.Blob{Data: audioData, MIMEType: "audio/wav"}},
			},
		},
	}
	config := t.buildGenerateConfig()
	return t.client.Models.GenerateContent(ctx, t.modelName, contents, config)
}

func (t *Transcriber) buildGenerateConfig() *genai.GenerateContentConfig {
	var tc *genai.ThinkingConfig
	if t.thinkingMode == "level" {
		tc = &genai.ThinkingConfig{ThinkingLevel: t.thinkingLevel}
	} else {
		tc = &genai.ThinkingConfig{ThinkingBudget: genai.Ptr[int32](0)}
	}

	cachedName := t.cacheNameByModel[t.modelName]
	if cachedName != "" {
		return &genai.GenerateContentConfig{
			CachedContent:  cachedName,
			ThinkingConfig: tc,
			Temperature:    genai.Ptr[float32](0.0),
		}
	}

	return &genai.GenerateContentConfig{
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{{Text: t.systemPrompt}},
		},
		ThinkingConfig: tc,
		Temperature:    genai.Ptr[float32](0.0),
	}
}

func (t *Transcriber) ensurePromptCache(ctx context.Context) {
	if !t.enablePromptCache {
		return
	}
	if _, ok := t.cacheNameByModel[t.modelName]; ok {
		return
	}

	cache, err := t.client.Caches.Create(ctx, t.modelName, &genai.CreateCachedContentConfig{
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{{Text: t.systemPrompt}},
		},
		TTL:         t.promptCacheTTL,
		DisplayName: "vibescribe-system-prompt-cache",
	})
	if err != nil {
		log.Printf("[Gemini] Prompt cache unavailable. system_instruction fallback を使用します: %v", err)
		return
	}
	if cache.Name != "" {
		t.cacheNameByModel[t.modelName] = cache.Name
		log.Printf("[Gemini] Prompt cache created: %s (model=%s)", cache.Name, t.modelName)
	}
}

func resolveThinkingLevel() genai.ThinkingLevel {
	raw := strings.TrimSpace(strings.ToLower(os.Getenv(ThinkingLevelEnvVar)))
	if raw == "" {
		raw = DefaultThinkingLevel
	}
	if level, ok := thinkingLevelMap[raw]; ok {
		return level
	}
	log.Printf("[Gemini] 無効な thinking level のため %s を使用します: %s", DefaultThinkingLevel, raw)
	return thinkingLevelMap[DefaultThinkingLevel]
}

func resolvePromptCacheEnabled() bool {
	raw := strings.TrimSpace(strings.ToLower(os.Getenv(EnablePromptCacheEnvVar)))
	switch raw {
	case "0", "false", "off", "no":
		return false
	default:
		return true
	}
}

func resolvePromptCacheTTL() time.Duration {
	raw := strings.TrimSpace(os.Getenv(PromptCacheTTLEnvVar))
	if raw == "" {
		return DefaultPromptCacheTTL
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		log.Printf("[Gemini] 無効な prompt cache TTL のためデフォルトを使用します: %s", raw)
		return DefaultPromptCacheTTL
	}
	return d
}
