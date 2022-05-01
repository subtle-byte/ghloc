package service

import "strings"

func filtered(s string, filter []string) bool {
	s = "^" + s + "$"
	filtered := false
	for _, pattern := range filter {
		if strings.HasPrefix(pattern, "!") {
			if strings.Contains(s, pattern[1:]) {
				filtered = false
			}
		} else {
			if strings.Contains(s, pattern) {
				filtered = true
			}
		}
	}
	return filtered
}
