package loc_count

import "strings"

func filtered(s string, filter *string) bool {
	s = "^" + s + "$"
	f := *filter
	f = strings.TrimSpace(f)
	filtered := strings.HasPrefix(f, "!")
	for _, pattern := range strings.Split(f, ",") {
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
