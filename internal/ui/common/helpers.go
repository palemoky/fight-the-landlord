// Package common provides shared utilities for the UI.
package common

// TruncateName truncates a player name to the specified maximum length.
func TruncateName(name string, maxLen int) string {
	runes := []rune(name)
	if len(runes) > maxLen {
		return string(runes[:maxLen-1]) + "…"
	}
	return name
}
