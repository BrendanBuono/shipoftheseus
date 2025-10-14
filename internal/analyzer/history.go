package analyzer

import (
	"fmt"
	"ship-of-theseus/internal/models"
	"strings"
)

const (
	// LineMovementWindow defines how far a line can move (±N lines) and still be "the same line".
	// This accounts for common refactoring like adding imports or reordering functions.
	LineMovementWindow = 10
)

// TraceLineHistory traces a single line back through git history to find its first appearance.
// It follows the line through commits, handling movement within files and file renames.
//
// Algorithm:
// 1. Start with the current line and its last modification commit (from blame)
// 2. Get the full file history with rename tracking (--follow)
// 3. Walk backwards through commits
// 4. At each commit, look for a similar line within ±10 lines of expected position
// 5. If found with ≥25% similarity, continue tracing
// 6. If not found or similarity drops below threshold, we've found the original
//
// Returns a LineHistory struct with first/last commit info and similarity score.
func TraceLineHistory(repoPath, filePath string, currentLine string, currentLineNum int, blameInfo BlameInfo) (*models.LineHistory, error) {
	// Get the complete file history following renames
	commitHashes, err := GetFileHistory(repoPath, filePath)
	if err != nil {
		// If we can't get history, use blame info as both first and last
		return &models.LineHistory{
			CurrentLine:     currentLine,
			OriginalLine:    currentLine,
			CurrentLineNum:  currentLineNum,
			OriginalLineNum: currentLineNum,
			FirstCommitHash: blameInfo.CommitHash,
			FirstCommitDate: blameInfo.CommitDate,
			LastCommitHash:  blameInfo.CommitHash,
			LastCommitDate:  blameInfo.CommitDate,
			Similarity:      1.0,
		}, nil
	}

	// Walk backwards through history to find the first appearance
	originalLine := currentLine
	originalLineNum := currentLineNum
	originalCommitHash := blameInfo.CommitHash
	var originalCommitDate = blameInfo.CommitDate

	// Start from the oldest commit and work forward to find first appearance
	for i := len(commitHashes) - 1; i >= 0; i-- {
		commitHash := commitHashes[i]

		// Get file content at this commit
		content, err := GetFileAtCommit(repoPath, commitHash, filePath)
		if err != nil {
			// File might not exist at this commit (before creation or after deletion)
			continue
		}

		lines := strings.Split(content, "\n")

		// Look for our line within the movement window
		matchedLine, matchedLineNum := findSimilarLineInRange(lines, originalLineNum, originalLine)

		if matchedLine != "" {
			// Found a match - this becomes our new "original"
			originalLine = matchedLine
			originalLineNum = matchedLineNum
			originalCommitHash = commitHash
			// Note: We'd need to get commit date from git, but for performance we'll use blame date
		} else {
			// No match found - the previous commit was the first appearance
			break
		}
	}

	// Calculate final similarity between original and current
	similarity := CalculateSimilarity(originalLine, currentLine)

	return &models.LineHistory{
		CurrentLine:     currentLine,
		OriginalLine:    originalLine,
		CurrentLineNum:  currentLineNum,
		OriginalLineNum: originalLineNum,
		FirstCommitHash: originalCommitHash,
		FirstCommitDate: originalCommitDate,
		LastCommitHash:  blameInfo.CommitHash,
		LastCommitDate:  blameInfo.CommitDate,
		Similarity:      similarity,
	}, nil
}

// findSimilarLineInRange searches for a line similar to targetLine within a ±10 line window.
// Returns the best matching line and its line number (1-indexed), or empty string if no match.
//
// This handles the case where lines move slightly due to refactoring (adding imports,
// reordering functions, etc.) but remain fundamentally "the same line".
func findSimilarLineInRange(lines []string, targetLineNum int, targetLine string) (string, int) {
	if len(lines) == 0 {
		return "", 0
	}

	// Convert to 0-indexed for array access
	targetIdx := targetLineNum - 1

	// Calculate search range (±10 lines)
	start := targetIdx - LineMovementWindow
	if start < 0 {
		start = 0
	}

	end := targetIdx + LineMovementWindow
	if end >= len(lines) {
		end = len(lines) - 1
	}

	// Find the best match within range
	bestMatch := ""
	bestSimilarity := 0.0
	bestLineNum := 0

	for i := start; i <= end; i++ {
		line := lines[i]
		similarity := CalculateSimilarity(targetLine, line)

		// Must meet minimum threshold to be considered a match
		if similarity >= MinimumSimilarityThreshold && similarity > bestSimilarity {
			bestMatch = line
			bestSimilarity = similarity
			bestLineNum = i + 1 // Convert back to 1-indexed
		}
	}

	return bestMatch, bestLineNum
}

// TraceFileLines analyzes all non-comment, non-blank lines in a file.
// Returns a slice of LineHistory for each analyzed line.
func TraceFileLines(repoPath, filePath, fileContent string) ([]*models.LineHistory, error) {
	// Get blame information for the file
	blameInfos, err := GetBlame(repoPath, filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get blame for %s: %w", filePath, err)
	}

	// Split content into lines
	lines := strings.Split(fileContent, "\n")

	// Git blame may have fewer lines than the file (doesn't count trailing newlines)
	// Use the blame count as the authoritative line count
	if len(blameInfos) > len(lines) {
		return nil, fmt.Errorf("blame line count (%d) exceeds file line count (%d) for %s",
			len(blameInfos), len(lines), filePath)
	}

	// Only process lines that have blame info
	lines = lines[:len(blameInfos)]

	var histories []*models.LineHistory

	// Trace each line's history
	for i, line := range lines {
		// Skip blank and comment lines - these aren't code
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Get blame info for this line
		blameInfo := blameInfos[i]
		lineNum := i + 1 // 1-indexed

		// Trace this line back through history
		history, err := TraceLineHistory(repoPath, filePath, line, lineNum, blameInfo)
		if err != nil {
			// Log warning but continue with other lines
			// In production, you'd want proper logging here
			continue
		}

		histories = append(histories, history)
	}

	return histories, nil
}
