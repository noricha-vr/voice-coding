package prompt

import (
	"bufio"
	"errors"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"strings"
)

// DictionaryPath returns the path to the user dictionary file.
func DictionaryPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = os.Getenv("HOME")
	}
	return filepath.Join(home, ".voicecode", "dictionary.txt")
}

// ParseDictionary reads the dictionary file and returns conversion XML and hint XML.
// Format:
//   - Lines with TAB: conversion entries (japanese\tenglish)
//   - Lines without TAB: hint words
//   - Lines starting with #: comments (skipped)
//   - Empty lines: skipped
//
// If file doesn't exist, returns empty strings (no error).
func ParseDictionary(path string) (conversionXML, hintXML string, err error) {
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", "", nil
		}
		return "", "", fmt.Errorf("opening dictionary: %w", err)
	}
	defer f.Close()

	var conversions []string
	var hints []string

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if parts := strings.SplitN(line, "\t", 2); len(parts) == 2 {
			japanese := strings.TrimSpace(parts[0])
			english := strings.TrimSpace(parts[1])
			if japanese != "" && english != "" {
				conversions = append(conversions, fmt.Sprintf(
					`<term japanese="%s" english="%s" context="always"/>`,
					html.EscapeString(japanese),
					html.EscapeString(english),
				))
			}
		} else {
			hints = append(hints, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return "", "", fmt.Errorf("reading dictionary: %w", err)
	}

	if len(conversions) > 0 {
		conversionXML = fmt.Sprintf(
			"<category name=\"ユーザー辞書（変換）\">\n%s\n</category>",
			strings.Join(conversions, "\n"),
		)
	}

	if len(hints) > 0 {
		hintXML = fmt.Sprintf(
			"<category name=\"ユーザー辞書（ヒント）\" type=\"hint\">\n<hint>%s</hint>\n<note>これらの単語はプログラミング文脈で頻繁に使用されます。音声認識結果にこれらの単語が含まれる可能性が高い場合、優先的に採用してください。</note>\n</category>",
			html.EscapeString(strings.Join(hints, ", ")),
		)
	}

	return conversionXML, hintXML, nil
}
