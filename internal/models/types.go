// Package models defines the core data structures for the Ship of Theseus analyzer.
// These structures represent the results of tracing code evolution through git history.
package models

import "time"

// CodebaseAnalysis represents the complete analysis results for a repository.
// It aggregates statistics across all analyzed files and includes historical snapshots.
type CodebaseAnalysis struct {
	TotalLines          int             // Total lines of code analyzed (excluding comments, blanks)
	OriginalLines       int             // Lines that are ≥25% similar to their first appearance
	AverageSimilarity   float64         // Mean similarity across all lines (0.0 to 1.0)
	FileAnalyses        []*FileAnalysis // Per-file detailed results
	HistoricalSnapshots []Snapshot      // Timeline of code evolution
}

// FileAnalysis represents the analysis results for a single file.
// It contains line-by-line history tracing and aggregated file metrics.
type FileAnalysis struct {
	Path          string         // Relative path from repository root
	TotalLines    int            // Total lines analyzed in this file
	OriginalLines int            // Lines that are ≥25% similar to original
	AvgSimilarity float64        // Mean similarity for this file's lines
	LineHistories []*LineHistory // Detailed history for each line
}

// LineHistory traces a single line from its first appearance to current state.
// It tracks the line's evolution through git history, measuring similarity.
type LineHistory struct {
	CurrentLine     string    // The line as it appears in current HEAD
	OriginalLine    string    // The line as it first appeared in history
	CurrentLineNum  int       // Line number in current file (1-indexed)
	OriginalLineNum int       // Line number in original commit (1-indexed)
	FirstCommitHash string    // Git commit hash where line first appeared
	FirstCommitDate time.Time // Date of first commit
	LastCommitHash  string    // Git commit hash of most recent modification
	LastCommitDate  time.Time // Date of most recent modification
	Similarity      float64   // Levenshtein similarity (0.0 to 1.0)
}

// Snapshot represents the estimated state of the codebase at a point in history.
// Used to generate the evolution timeline graph showing how originality changes over time.
type Snapshot struct {
	CommitHash  string    // Git commit hash for this snapshot
	Date        time.Time // Commit date
	OriginalPct float64   // Estimated percentage of original code remaining
}
