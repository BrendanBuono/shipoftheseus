package analyzer

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"ship-of-theseus/internal/filter"
	"ship-of-theseus/internal/models"
	"strings"
	"sync"
)

// AnalyzeRepository performs a complete Ship of Theseus analysis on a git repository.
// It processes files in parallel and returns aggregated statistics about code originality.
//
// Parameters:
//   - repoPath: Absolute path to the git repository
//   - numWorkers: Number of parallel workers (use runtime.NumCPU() for default)
//
// Returns:
//   - CodebaseAnalysis with complete metrics and per-file breakdowns
func AnalyzeRepository(repoPath string, numWorkers int) (*models.CodebaseAnalysis, error) {
	// Validate repository exists
	if _, err := os.Stat(filepath.Join(repoPath, ".git")); os.IsNotExist(err) {
		return nil, fmt.Errorf("not a git repository: %s", repoPath)
	}

	// Find all files in repository
	files, err := findAllFiles(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to find files: %w", err)
	}

	// Filter out files we should skip (binary, generated, vendor)
	var filesToAnalyze []string
	for _, file := range files {
		relPath, err := filepath.Rel(repoPath, file)
		if err != nil {
			continue
		}

		if !filter.ShouldSkipFile(relPath) {
			filesToAnalyze = append(filesToAnalyze, relPath)
		}
	}

	if len(filesToAnalyze) == 0 {
		return nil, fmt.Errorf("no files to analyze after filtering")
	}

	fmt.Printf("Analyzing %d files with %d workers...\n", len(filesToAnalyze), numWorkers)

	// Process files in parallel using worker pool
	fileAnalyses := processFilesParallel(repoPath, filesToAnalyze, numWorkers)

	// Aggregate results
	analysis := aggregateResults(fileAnalyses)

	return analysis, nil
}

// findAllFiles recursively finds all files in a directory.
func findAllFiles(root string) ([]string, error) {
	var files []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip .git directory itself
		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}

		// Only include regular files
		if !info.IsDir() {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

// processFilesParallel processes files using a worker pool for parallelization.
// This is critical for performance on large repositories.
func processFilesParallel(repoPath string, files []string, numWorkers int) []*models.FileAnalysis {
	// Create channels for work distribution
	workChan := make(chan string, len(files))
	resultChan := make(chan *models.FileAnalysis, len(files))

	// WaitGroup to track worker completion
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker(repoPath, workChan, resultChan, &wg)
	}

	// Send work to workers
	for _, file := range files {
		workChan <- file
	}
	close(workChan)

	// Wait for all workers to finish, then close result channel
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	var results []*models.FileAnalysis
	for result := range resultChan {
		if result != nil {
			results = append(results, result)
			fmt.Printf("Completed: %s (%d lines, %.1f%% original)\n",
				result.Path, result.TotalLines, float64(result.OriginalLines)/float64(result.TotalLines)*100)
		}
	}

	return results
}

// worker processes files from the work channel and sends results to result channel.
func worker(repoPath string, workChan <-chan string, resultChan chan<- *models.FileAnalysis, wg *sync.WaitGroup) {
	defer wg.Done()

	for filePath := range workChan {
		analysis, err := analyzeFile(repoPath, filePath)
		if err != nil {
			// Log error but continue processing other files
			fmt.Printf("Warning: Failed to analyze %s: %v\n", filePath, err)
			continue
		}

		resultChan <- analysis
	}
}

// analyzeFile performs a complete analysis of a single file.
func analyzeFile(repoPath, filePath string) (*models.FileAnalysis, error) {
	// Read file content
	fullPath := filepath.Join(repoPath, filePath)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Split into lines and filter out comments/blanks
	lines := strings.Split(string(content), "\n")
	var codeLines []string
	var codeLinesNums []int

	for i, line := range lines {
		if !filter.IsBlankOrComment(line, filePath) {
			codeLines = append(codeLines, line)
			codeLinesNums = append(codeLinesNums, i+1)
		}
	}

	// Skip files with no code lines
	if len(codeLines) == 0 {
		return nil, fmt.Errorf("no code lines found (all comments/blanks)")
	}

	// Trace line histories
	histories, err := TraceFileLines(repoPath, filePath, string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to trace lines: %w", err)
	}

	// Calculate file metrics
	totalLines := len(histories)
	originalLines := 0
	totalSimilarity := 0.0

	for _, history := range histories {
		if IsOriginal(history.Similarity) {
			originalLines++
		}
		totalSimilarity += history.Similarity
	}

	avgSimilarity := 0.0
	if totalLines > 0 {
		avgSimilarity = totalSimilarity / float64(totalLines)
	}

	return &models.FileAnalysis{
		Path:          filePath,
		TotalLines:    totalLines,
		OriginalLines: originalLines,
		AvgSimilarity: avgSimilarity,
		LineHistories: histories,
	}, nil
}

// aggregateResults combines individual file analyses into repository-wide statistics.
func aggregateResults(fileAnalyses []*models.FileAnalysis) *models.CodebaseAnalysis {
	totalLines := 0
	originalLines := 0
	totalSimilarity := 0.0

	for _, fa := range fileAnalyses {
		totalLines += fa.TotalLines
		originalLines += fa.OriginalLines
		totalSimilarity += fa.AvgSimilarity * float64(fa.TotalLines)
	}

	avgSimilarity := 0.0
	if totalLines > 0 {
		avgSimilarity = totalSimilarity / float64(totalLines)
	}

	return &models.CodebaseAnalysis{
		TotalLines:          totalLines,
		OriginalLines:       originalLines,
		AverageSimilarity:   avgSimilarity,
		FileAnalyses:        fileAnalyses,
		HistoricalSnapshots: []models.Snapshot{}, // Will be filled by snapshot generation
	}
}

// GetDefaultWorkerCount returns the recommended number of workers (number of CPUs).
func GetDefaultWorkerCount() int {
	return runtime.NumCPU()
}
