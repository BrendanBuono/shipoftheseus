// Ship of Theseus - A git codebase evolution analyzer
//
// This tool traces each line of code from its first appearance to its current state,
// measuring similarity to answer the philosophical question: "Is this still the same codebase?"
//
// Named after the ancient paradox: If you replace every plank of a ship over time,
// is it still the same ship?
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"ship-of-theseus/internal/analyzer"
	"ship-of-theseus/internal/visualizer"
)

const version = "1.0.0"

func main() {
	// Define command-line flags
	var (
		repoPath   = flag.String("path", ".", "Path to git repository")
		numWorkers = flag.Int("workers", analyzer.GetDefaultWorkerCount(), "Number of parallel workers")
		sampleRate = flag.Int("sample", 50, "Sample every Nth commit for history timeline")
		showVersion = flag.Bool("version", false, "Show version information")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Ship of Theseus v%s - Codebase Evolution Analyzer

Usage:
  ship-of-theseus [options]

Options:
`, version)
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, `
Description:
  Analyzes a git repository to determine how much "original" code remains.
  Traces each line from its first appearance to its current state, measuring
  similarity using Levenshtein distance. Lines with ≥25%% similarity are
  considered "original"; lines that have changed more are "completely different."

  Like the ancient Ship of Theseus paradox: if every line of code is eventually
  modified, is it still the same codebase?

Examples:
  # Analyze current directory
  ship-of-theseus

  # Analyze specific repository with 8 workers
  ship-of-theseus --path /path/to/repo --workers 8

  # Use coarser sampling for faster analysis
  ship-of-theseus --sample 100

Performance:
  - Small repos (<100 files): <30 seconds
  - Medium repos (1K-5K files): <5 minutes
  - Large repos (>10K files): <30 minutes

  Increase --workers for faster analysis on multi-core systems.
  Increase --sample to reduce analysis time (trades accuracy for speed).

For more information: https://github.com/yourusername/ship-of-theseus
`)
	}

	flag.Parse()

	// Handle version flag
	if *showVersion {
		fmt.Printf("Ship of Theseus v%s\n", version)
		os.Exit(0)
	}

	// Validate and resolve repository path
	absPath, err := filepath.Abs(*repoPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Invalid path: %v\n", err)
		os.Exit(1)
	}

	// Check if directory exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: Directory does not exist: %s\n", absPath)
		os.Exit(1)
	}

	// Check if it's a git repository
	gitDir := filepath.Join(absPath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: Not a git repository: %s\n", absPath)
		fmt.Fprintf(os.Stderr, "       (no .git directory found)\n")
		os.Exit(1)
	}

	// Validate parameters
	if *numWorkers < 1 {
		fmt.Fprintf(os.Stderr, "Error: --workers must be at least 1\n")
		os.Exit(1)
	}

	if *sampleRate < 1 {
		fmt.Fprintf(os.Stderr, "Error: --sample must be at least 1\n")
		os.Exit(1)
	}

	// Print startup message
	fmt.Printf("🚢 Ship of Theseus v%s\n", version)
	fmt.Printf("Analyzing repository: %s\n", absPath)
	fmt.Printf("Workers: %d | Sample rate: every %d commits\n\n", *numWorkers, *sampleRate)

	// Run the analysis
	analysis, err := analyzer.AnalyzeRepository(absPath, *numWorkers)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\nError during analysis: %v\n", err)
		os.Exit(1)
	}

	// Generate historical snapshots
	fmt.Println("\nGenerating historical timeline...")
	if err := analyzer.AddSnapshotsToAnalysis(analysis, absPath, *sampleRate); err != nil {
		// Non-fatal: continue without snapshots
		fmt.Printf("Warning: Could not generate historical timeline: %v\n", err)
	}

	// Display results
	visualizer.Display(analysis)
}
