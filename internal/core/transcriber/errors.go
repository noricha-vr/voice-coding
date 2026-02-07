package transcriber

import (
	"regexp"
	"strings"
)

var httpErrorCodePattern = regexp.MustCompile(`\b(429|500|502|503|504)\b`)

// IsModelNotFound returns true if the error indicates the model was not found.
func IsModelNotFound(err error) bool {
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "is not found for api version") ||
		(strings.Contains(msg, "404") && strings.Contains(msg, "models/") && strings.Contains(msg, "not found"))
}

// IsTransient returns true if the error is a transient API error that may succeed on retry.
func IsTransient(err error) bool {
	msg := strings.ToLower(err.Error())
	patterns := []string{
		"deadline expired",
		"timed out",
		"timeout",
		"temporarily unavailable",
		"service unavailable",
	}
	for _, p := range patterns {
		if strings.Contains(msg, p) {
			return true
		}
	}
	return httpErrorCodePattern.MatchString(msg) || strings.Contains(msg, "too many requests")
}

// IsThinkingUnsupported returns true if the error indicates thinking level is not supported.
func IsThinkingUnsupported(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "thinking level is not supported")
}

// IsCachedContentError returns true if the error indicates a cached content problem.
func IsCachedContentError(err error) bool {
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "cachedcontent not found") ||
		(strings.Contains(msg, "permission_denied") && strings.Contains(msg, "cached"))
}
