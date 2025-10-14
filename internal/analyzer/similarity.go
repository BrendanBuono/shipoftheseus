// Package analyzer implements the core Ship of Theseus analysis logic.
// It traces lines through git history, calculates similarity, and generates metrics.
package analyzer

import (
	"strings"

	"github.com/agnivade/levenshtein"
)

const (
	// MinimumSimilarityThreshold is the cutoff below which lines are considered completely different.
	// At 25%, lines share only basic structural elements and are essentially rewritten.
	MinimumSimilarityThreshold = 0.25
)

// CalculateSimilarity compares two strings using Levenshtein distance and returns
// a similarity score between 0.0 (completely different) and 1.0 (identical).
//
// The algorithm:
// 1. Trims whitespace from both strings (formatting changes don't affect similarity)
// 2. Returns 1.0 for exact matches (fast path)
// 3. Returns 0.0 if either string is empty
// 4. Computes Levenshtein distance (minimum edits: insertions, deletions, substitutions)
// 5. Normalizes by the length of the longer string
//
// Example:
//   CalculateSimilarity("func add(a, b)", "func add(x, y)") → ~0.85
//   CalculateSimilarity("hello world", "hello") → ~0.45
func CalculateSimilarity(original, current string) float64 {
	// Normalize whitespace - formatting changes shouldn't affect similarity
	original = strings.TrimSpace(original)
	current = strings.TrimSpace(current)

	// Fast path: exact match
	if original == current {
		return 1.0
	}

	// Empty strings are completely different from non-empty strings
	if original == "" || current == "" {
		return 0.0
	}

	// Compute Levenshtein distance (number of edits needed to transform one into the other)
	distance := levenshtein.ComputeDistance(original, current)

	// Normalize by the length of the longer string
	// This gives us a percentage: how much of the string remained unchanged?
	maxLen := max(len(original), len(current))
	similarity := 1.0 - (float64(distance) / float64(maxLen))

	// Clamp to [0, 1] range (should always be in range, but be defensive)
	if similarity < 0.0 {
		return 0.0
	}
	if similarity > 1.0 {
		return 1.0
	}

	return similarity
}

// IsOriginal determines if a line should be considered "original" based on similarity.
// Returns true if similarity meets the minimum threshold (25%).
func IsOriginal(similarity float64) bool {
	return similarity >= MinimumSimilarityThreshold
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
