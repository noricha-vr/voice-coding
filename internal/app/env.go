package app

import (
	"os"
	"strings"
)

func envBoolDefaultTrue(name string) bool {
	raw := strings.TrimSpace(strings.ToLower(os.Getenv(name)))
	switch raw {
	case "0", "false", "off", "no":
		return false
	default:
		return true
	}
}
