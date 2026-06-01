package stdlib

import (
	"fmt"
	"regexp"
)

// Regular expression utilities for Kylix

// TRegex wraps Go's regexp.Regexp for Pascal-style regex operations
type TRegex struct {
	re      *regexp.Regexp
	pattern string
}

// RegexCompile compiles a regular expression pattern
func RegexCompile(pattern string) (*TRegex, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("RegexCompile: %w", err)
	}
	return &TRegex{re: re, pattern: pattern}, nil
}

// RegexMustCompile compiles a regex pattern, panics on error
func RegexMustCompile(pattern string) *TRegex {
	re := regexp.MustCompile(pattern)
	return &TRegex{re: re, pattern: pattern}
}

// Match checks if the string matches the pattern
func (r *TRegex) Match(s string) bool {
	return r.re.MatchString(s)
}

// Find returns the first match
func (r *TRegex) Find(s string) string {
	return r.re.FindString(s)
}

// FindAll returns all matches
func (r *TRegex) FindAll(s string) []string {
	return r.re.FindAllString(s, -1)
}

// FindN returns up to n matches
func (r *TRegex) FindN(s string, n int) []string {
	return r.re.FindAllString(s, n)
}

// Replace replaces all matches with the replacement
func (r *TRegex) Replace(s, replacement string) string {
	return r.re.ReplaceAllString(s, replacement)
}

// ReplaceFirst replaces the first match
func (r *TRegex) ReplaceFirst(s, replacement string) string {
	loc := r.re.FindStringIndex(s)
	if loc == nil {
		return s
	}
	return s[:loc[0]] + replacement + s[loc[1]:]
}

// Split splits the string by the pattern
func (r *TRegex) Split(s string) []string {
	return r.re.Split(s, -1)
}

// SplitN splits into at most n parts
func (r *TRegex) SplitN(s string, n int) []string {
	return r.re.Split(s, n)
}

// Groups returns the capture groups for the first match
func (r *TRegex) Groups(s string) []string {
	return r.re.FindStringSubmatch(s)
}

// GroupsAll returns capture groups for all matches
func (r *TRegex) GroupsAll(s string) [][]string {
	return r.re.FindAllStringSubmatch(s, -1)
}

// Pattern returns the original pattern string
func (r *TRegex) Pattern() string {
	return r.pattern
}

// Convenience functions (no need to compile first)

// RegexMatch checks if a string matches a pattern
func RegexMatch(pattern, s string) bool {
	matched, err := regexp.MatchString(pattern, s)
	if err != nil {
		return false
	}
	return matched
}

// RegexFind returns the first match of a pattern in a string
func RegexFind(pattern, s string) string {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return ""
	}
	return re.FindString(s)
}

// RegexReplace replaces all matches of a pattern
func RegexReplace(pattern, s, replacement string) string {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return s
	}
	return re.ReplaceAllString(s, replacement)
}

// RegexSplit splits a string by a pattern
func RegexSplit(pattern, s string) []string {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return []string{s}
	}
	return re.Split(s, -1)
}

// Common pattern helpers

// IsEmail checks if a string looks like an email address
func IsEmail(s string) bool {
	return RegexMatch(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`, s)
}

// IsURL checks if a string looks like a URL
func IsURL(s string) bool {
	return RegexMatch(`^https?://[^\s/$.?#].[^\s]*$`, s)
}

// IsNumeric checks if a string contains only digits
func IsNumeric(s string) bool {
	return RegexMatch(`^\d+$`, s)
}

// IsAlpha checks if a string contains only letters
func IsAlpha(s string) bool {
	return RegexMatch(`^[a-zA-Z]+$`, s)
}

// IsAlphaNumeric checks if a string contains only letters and digits
func IsAlphaNumeric(s string) bool {
	return RegexMatch(`^[a-zA-Z0-9]+$`, s)
}

// IsIP checks if a string looks like an IPv4 address
func IsIP(s string) bool {
	return RegexMatch(`^(\d{1,3}\.){3}\d{1,3}$`, s)
}

// ExtractNumbers extracts all numbers from a string
func ExtractNumbers(s string) []string {
	return RegexFind2(`\d+`, s)
}

// RegexFind2 is a helper that finds all matches
func RegexFind2(pattern, s string) []string {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil
	}
	return re.FindAllString(s, -1)
}
