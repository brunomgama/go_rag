package common

import (
	"strings"
)

func IsNilValue(value any) bool {
	if value == nil {
		return true
	}
	return false
}

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func Snippet(s string, n int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func Clamp(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
