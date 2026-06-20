// suggestions.go — Levenshtein-based spell correction and fix hints.
package compiler

import (
	"kylix/pkg/i18n"
	"strings"
)

// NearestName returns the closest name in candidates to target
// if the edit distance is <= maxDist. Returns "" if nothing is close enough.
func NearestName(target string, candidates []string, maxDist int) string {
	best := ""
	bestDist := maxDist + 1
	for _, c := range candidates {
		d := levenshtein(strings.ToLower(target), strings.ToLower(c))
		if d < bestDist {
			bestDist = d
			best = c
		}
	}
	if bestDist <= maxDist {
		return best
	}
	return ""
}

// levenshtein computes the edit distance between two strings.
func levenshtein(a, b string) int {
	ra, rb := []rune(a), []rune(b)
	la, lb := len(ra), len(rb)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}

	// Two-row DP to keep memory O(min(la,lb))
	if la < lb {
		ra, rb = rb, ra
		la, lb = lb, la
	}

	prev := make([]int, lb+1)
	curr := make([]int, lb+1)
	for j := 0; j <= lb; j++ {
		prev[j] = j
	}
	for i := 1; i <= la; i++ {
		curr[0] = i
		for j := 1; j <= lb; j++ {
			cost := 1
			if ra[i-1] == rb[j-1] {
				cost = 0
			}
			del := curr[j-1] + 1
			ins := prev[j] + 1
			sub := prev[j-1] + cost
			curr[j] = min3(del, ins, sub)
		}
		prev, curr = curr, prev
	}
	return prev[lb]
}

func min3(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

// typeConversionHint returns a conversion function suggestion when assigning
// the wrong literal type to a typed variable.
// declaredType: the variable's declared type, valueLiteralType: "string"/"integer"
func typeConversionHint(declaredType, valueLiteralType string) string {
	norm := strings.ToLower(declaredType)
	switch valueLiteralType {
	case "string":
		switch norm {
		case "integer", "int64":
			return i18n.Hint("KLX101_str_to_int")
		case "real", "double", "float64":
			return i18n.Hint("KLX101_str_to_float")
		case "boolean":
			return i18n.Hint("KLX101_str_to_bool")
		}
	case "integer":
		if norm == "string" {
			return i18n.Hint("KLX101_int_to_str")
		}
	}
	return ""
}
