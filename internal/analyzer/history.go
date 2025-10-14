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
// Simplified algorithm: Get the line from the file's first commit and compare to current.
//
// Algorithm:
// 1. Get the file's first commit (oldest in history)
// 2. Get the line at the same position (±10 lines) in that first commit
// 3. Compare first commit's line to current line
// 4. Calculate similarity
//
// Returns a LineHistory struct with first/last commit info and similarity score.
func TraceLineHistory(repoPath, filePath string, currentLine string, currentLineNum int, blameInfo BlameInfo) (*models.LineHistory, error) {
	// Get the complete file history following renames
	commitHashes, err := GetFileHistory(repoPath, filePath)
	if err != nil || len(commitHashes) == 0 {
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

	// Get the FIRST (oldest) commit where this file existed
	firstCommitHash := commitHashes[len(commitHashes)-1]

	// Get file content at first commit
	firstContent, err := GetFileAtCommit(repoPath, firstCommitHash, filePath)
	if err != nil {
		// File didn't exist in first commit - treat current as original
		return &models.LineHistory{
			CurrentLine:     currentLine,
			OriginalLine:    currentLine,
			CurrentLineNum:  currentLineNum,
			OriginalLineNum: currentLineNum,
			FirstCommitHash: firstCommitHash,
			FirstCommitDate: blameInfo.CommitDate,
			LastCommitHash:  blameInfo.CommitHash,
			LastCommitDate:  blameInfo.CommitDate,
			Similarity:      1.0,
		}, nil
	}

	firstLines := strings.Split(firstContent, "\n")

	// Look for a similar line in the first commit within ±10 lines of current position
	originalLine, originalLineNum := findSimilarLineInRange(firstLines, currentLineNum, currentLine)

	if originalLine == "" {
		// No similar line found in first commit - this is a new line
		originalLine = ""
		originalLineNum = currentLineNum
	}

	// Calculate similarity between first and current
	similarity := CalculateSimilarity(originalLine, currentLine)

	return &models.LineHistory{
		CurrentLine:     currentLine,
		OriginalLine:    originalLine,
		CurrentLineNum:  currentLineNum,
		OriginalLineNum: originalLineNum,
		FirstCommitHash: firstCommitHash,
		FirstCommitDate: blameInfo.CommitDate, // Simplified: use blame date
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
