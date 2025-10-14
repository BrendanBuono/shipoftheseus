package filter

import (
	"path/filepath"
	"regexp"
	"strings"
)

// commentPatterns maps file extensions to their comment detection patterns.
// Each language can have multiple patterns for different comment styles.
var commentPatterns = map[string][]*regexp.Regexp{
	// C-style languages (C, C++, Go, Java, JavaScript, etc.)
	".go":   {regexp.MustCompile(`^\s*//`), regexp.MustCompile(`^\s*/\*`)},
	".c":    {regexp.MustCompile(`^\s*//`), regexp.MustCompile(`^\s*/\*`)},
	".cpp":  {regexp.MustCompile(`^\s*//`), regexp.MustCompile(`^\s*/\*`)},
	".h":    {regexp.MustCompile(`^\s*//`), regexp.MustCompile(`^\s*/\*`)},
	".hpp":  {regexp.MustCompile(`^\s*//`), regexp.MustCompile(`^\s*/\*`)},
	".java": {regexp.MustCompile(`^\s*//`), regexp.MustCompile(`^\s*/\*`)},
	".js":   {regexp.MustCompile(`^\s*//`), regexp.MustCompile(`^\s*/\*`)},
	".ts":   {regexp.MustCompile(`^\s*//`), regexp.MustCompile(`^\s*/\*`)},
	".jsx":  {regexp.MustCompile(`^\s*//`), regexp.MustCompile(`^\s*/\*`)},
	".tsx":  {regexp.MustCompile(`^\s*//`), regexp.MustCompile(`^\s*/\*`)},
	".cs":   {regexp.MustCompile(`^\s*//`), regexp.MustCompile(`^\s*/\*`)},
	".swift": {regexp.MustCompile(`^\s*//`), regexp.MustCompile(`^\s*/\*`)},
	".kt":    {regexp.MustCompile(`^\s*//`), regexp.MustCompile(`^\s*/\*`)},
	".scala": {regexp.MustCompile(`^\s*//`), regexp.MustCompile(`^\s*/\*`)},
	".rs":    {regexp.MustCompile(`^\s*//`), regexp.MustCompile(`^\s*/\*`)},
	".php":   {regexp.MustCompile(`^\s*//`), regexp.MustCompile(`^\s*/\*`), regexp.MustCompile(`^\s*#`)},

	// Python and similar
	".py":  {regexp.MustCompile(`^\s*#`)},
	".rb":  {regexp.MustCompile(`^\s*#`)},
	".sh":  {regexp.MustCompile(`^\s*#`)},
	".bash": {regexp.MustCompile(`^\s*#`)},
	".zsh":  {regexp.MustCompile(`^\s*#`)},
	".fish": {regexp.MustCompile(`^\s*#`)},
	".pl":   {regexp.MustCompile(`^\s*#`)},
	".pm":   {regexp.MustCompile(`^\s*#`)},
	".r":    {regexp.MustCompile(`^\s*#`)},
	".yaml": {regexp.MustCompile(`^\s*#`)},
	".yml":  {regexp.MustCompile(`^\s*#`)},
	".toml": {regexp.MustCompile(`^\s*#`)},
	".conf": {regexp.MustCompile(`^\s*#`)},
	".ini":  {regexp.MustCompile(`^\s*#|^\s*;`)},

	// SQL
	".sql": {regexp.MustCompile(`^\s*--`), regexp.MustCompile(`^\s*/\*`)},

	// Lua
	".lua": {regexp.MustCompile(`^\s*--`)},

	// Lisp family
	".el":  {regexp.MustCompile(`^\s*;`)},
	".lisp": {regexp.MustCompile(`^\s*;`)},
	".clj":  {regexp.MustCompile(`^\s*;`)},

	// HTML/XML
	".html": {regexp.MustCompile(`^\s*<!--`)},
	".xml":  {regexp.MustCompile(`^\s*<!--`)},
	".svg":  {regexp.MustCompile(`^\s*<!--`)},

	// CSS and variants
	".css":  {regexp.MustCompile(`^\s*/\*`)},
	".scss": {regexp.MustCompile(`^\s*//`), regexp.MustCompile(`^\s*/\*`)},
	".sass": {regexp.MustCompile(`^\s*//`)},
	".less": {regexp.MustCompile(`^\s*//`), regexp.MustCompile(`^\s*/\*`)},

	// Other languages
	".vim":  {regexp.MustCompile(`^\s*"`)},
	".tex":  {regexp.MustCompile(`^\s*%`)},
	".m":    {regexp.MustCompile(`^\s*%`)}, // MATLAB/Octave
	".erl":  {regexp.MustCompile(`^\s*%`)}, // Erlang
	".ex":   {regexp.MustCompile(`^\s*#`)}, // Elixir
	".exs":  {regexp.MustCompile(`^\s*#`)},
	".hs":   {regexp.MustCompile(`^\s*--`)}, // Haskell
	".elm":  {regexp.MustCompile(`^\s*--`)},
	".ml":   {regexp.MustCompile(`^\s*\(\*`)}, // OCaml
}

// IsBlankOrComment determines if a line should be skipped because it's blank or comment-only.
// It uses the file extension to determine which comment patterns to check.
func IsBlankOrComment(line string, filePath string) bool {
	// Trim whitespace for checking
	trimmed := strings.TrimSpace(line)

	// Skip blank lines
	if trimmed == "" {
		return true
	}

	// Get file extension
	ext := strings.ToLower(filepath.Ext(filePath))

	// Get comment patterns for this file type
	patterns, exists := commentPatterns[ext]
	if !exists {
		// Unknown file type - only skip if blank
		return false
	}

	// Check if line matches any comment pattern
	for _, pattern := range patterns {
		if pattern.MatchString(line) {
			return true
		}
	}

	return false
}

// IsBlank checks if a line contains only whitespace.
func IsBlank(line string) bool {
	return strings.TrimSpace(line) == ""
}

// StripComments removes comment-only lines from a slice of lines.
// Returns a new slice containing only non-comment, non-blank lines.
func StripComments(lines []string, filePath string) []string {
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		if !IsBlankOrComment(line, filePath) {
			result = append(result, line)
		}
	}
	return result
}

// CountCodeLines counts the number of non-blank, non-comment lines in the input.
func CountCodeLines(lines []string, filePath string) int {
	count := 0
	for _, line := range lines {
		if !IsBlankOrComment(line, filePath) {
			count++
		}
	}
	return count
}
