package recommender

import (
	"sort"
	"strings"
)

func splitClean(s, sep string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, sep)
	var clean []string
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			clean = append(clean, t)
		}
	}
	return clean
}

func normaliseName(s string) string {
	words := strings.Fields(strings.ToLower(s))
	sort.Strings(words)
	return strings.Join(words, " ")
}

func intersectionCount(a, b []string) int {
	m := make(map[string]bool, len(a))
	for _, item := range a {
		m[normaliseName(item)] = true
	}
	count := 0
	for _, item := range b {
		if m[normaliseName(item)] {
			count++
		}
	}
	return count
}
