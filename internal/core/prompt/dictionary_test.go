package prompt

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseDictionaryBothConversionAndHint(t *testing.T) {
	path := filepath.Join("testdata", "dictionary_test.txt")

	conv, hint, err := ParseDictionary(path)
	if err != nil {
		t.Fatalf("ParseDictionary() error: %v", err)
	}

	// Check conversion XML
	if !strings.Contains(conv, `<category name="ユーザー辞書（変換）">`) {
		t.Error("conversion XML missing category tag")
	}
	if !strings.Contains(conv, `japanese="クロードコード"`) {
		t.Error("conversion XML missing クロードコード entry")
	}
	if !strings.Contains(conv, `english="Claude Code"`) {
		t.Error("conversion XML missing Claude Code entry")
	}
	if !strings.Contains(conv, `japanese="リアクト"`) {
		t.Error("conversion XML missing リアクト entry")
	}
	if !strings.Contains(conv, `english="Next.js"`) {
		t.Error("conversion XML missing Next.js entry")
	}

	// Check hint XML
	if !strings.Contains(hint, `<category name="ユーザー辞書（ヒント）" type="hint">`) {
		t.Error("hint XML missing category tag")
	}
	if !strings.Contains(hint, "haiku") {
		t.Error("hint XML missing haiku")
	}
	if !strings.Contains(hint, "gemini") {
		t.Error("hint XML missing gemini")
	}
	if !strings.Contains(hint, "typescript") {
		t.Error("hint XML missing typescript")
	}
}

func TestParseDictionaryConversionsOnly(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "dict.txt")
	content := "リアクト\tReact\nネクスト\tNext.js\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}

	conv, hint, err := ParseDictionary(path)
	if err != nil {
		t.Fatalf("ParseDictionary() error: %v", err)
	}

	if conv == "" {
		t.Error("expected non-empty conversion XML")
	}
	if hint != "" {
		t.Errorf("expected empty hint XML, got %q", hint)
	}
}

func TestParseDictionaryHintsOnly(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "dict.txt")
	content := "haiku\ngemini\ntypescript\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}

	conv, hint, err := ParseDictionary(path)
	if err != nil {
		t.Fatalf("ParseDictionary() error: %v", err)
	}

	if conv != "" {
		t.Errorf("expected empty conversion XML, got %q", conv)
	}
	if hint == "" {
		t.Error("expected non-empty hint XML")
	}
	if !strings.Contains(hint, "haiku, gemini, typescript") {
		t.Errorf("hint XML should contain comma-separated words, got %q", hint)
	}
}

func TestParseDictionaryCommentsAndEmptyLines(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "dict.txt")
	content := "# コメント\n\n# もう一つのコメント\nリアクト\tReact\n\n# 末尾コメント\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}

	conv, hint, err := ParseDictionary(path)
	if err != nil {
		t.Fatalf("ParseDictionary() error: %v", err)
	}

	if !strings.Contains(conv, `japanese="リアクト"`) {
		t.Error("conversion XML missing リアクト entry")
	}
	if strings.Contains(conv, "コメント") {
		t.Error("comments should not appear in output")
	}
	if hint != "" {
		t.Errorf("expected empty hint XML, got %q", hint)
	}
}

func TestParseDictionaryFileNotFound(t *testing.T) {
	conv, hint, err := ParseDictionary("/nonexistent/path/dict.txt")
	if err != nil {
		t.Fatalf("ParseDictionary() should not error on missing file, got: %v", err)
	}
	if conv != "" {
		t.Errorf("expected empty conversion XML, got %q", conv)
	}
	if hint != "" {
		t.Errorf("expected empty hint XML, got %q", hint)
	}
}

func TestParseDictionaryHTMLEscape(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "dict.txt")
	content := "テスト&アンド\tTest&And\n<script>\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}

	conv, hint, err := ParseDictionary(path)
	if err != nil {
		t.Fatalf("ParseDictionary() error: %v", err)
	}

	if !strings.Contains(conv, "テスト&amp;アンド") {
		t.Error("japanese value should have & escaped")
	}
	if !strings.Contains(conv, "Test&amp;And") {
		t.Error("english value should have & escaped")
	}

	if !strings.Contains(hint, "&lt;script&gt;") {
		t.Errorf("hint should have < and > escaped, got %q", hint)
	}
}
