# Project: Ship of Theseus - Git Codebase Evolution Analyzer

## Context & Philosophy
Build a CLI tool in Go that analyzes a git repository to determine how much of the "original" codebase remains, inspired by the Ship of Theseus philosophical paradox. The tool traces each line of code from its first appearance to its current state, measuring similarity to determine if it's still "the same line." This creates a fascinating metric: what percentage of a codebase is truly "original"?

## Development Workflow: Version Control Strategy

### CRITICAL: Git Commit Strategy
**You MUST commit your work to git after completing each logical group of changes.** This creates a meaningful version history of the tool's development.

#### Commit After Each of These Steps:
1. ✅ **Initial setup**: `git init`, create `go.mod`, add dependencies
   - Commit: "Initial project setup with dependencies"

2. ✅ **Data structures**: Create `internal/models/types.go` with all structs
   - Commit: "Add core data structures and models"

3. ✅ **File filtering**: Implement `internal/filter/files.go` and `internal/filter/comments.go`
   - Commit: "Implement file and comment filtering logic"

4. ✅ **Similarity calculation**: Implement `internal/analyzer/similarity.go`
   - Commit: "Add Levenshtein distance similarity calculation"

5. ✅ **Git CLI integration**: Implement `internal/analyzer/blame.go`
   - Commit: "Add git CLI wrapper for blame and file operations"

6. ✅ **Line history tracing**: Implement `internal/analyzer/history.go`
   - Commit: "Implement line history tracing with rename detection"

7. ✅ **Main analyzer**: Implement `internal/analyzer/analyzer.go` with parallelization
   - Commit: "Add main analyzer with parallel processing"

8. ✅ **Historical snapshots**: Implement `internal/analyzer/snapshots.go`
   - Commit: "Add historical timeline generation with heuristic estimation"

9. ✅ **Visualization**: Implement `internal/visualizer/graph.go`
   - Commit: "Add terminal visualization and output formatting"

10. ✅ **CLI entry point**: Implement `main.go` with flag parsing
    - Commit: "Add CLI entry point and command-line interface"

11. ✅ **Documentation**: Create README.md with usage instructions
    - Commit: "Add comprehensive README with examples"

12. ✅ **Bug fixes and refinements**: As you test and fix issues
    - Commit each fix separately with descriptive messages

#### Commit Message Guidelines:
- Use imperative mood: "Add", "Implement", "Fix", not "Added" or "Adding"
- Be descriptive but concise
- Reference the component/file being changed
- Examples:
  - ✅ "Add parallel file processing with worker pool"
  - ✅ "Fix line matching in rename detection"
  - ✅ "Optimize memory usage for large repositories"
  - ❌ "Update code"
  - ❌ "Changes"

#### Before Each Commit:
```bash
# Stage specific files related to the change
git add internal/analyzer/blame.go

# Commit with descriptive message
git commit -m "Add git CLI wrapper for blame operations"
```

This creates a clean, professional git history that demonstrates the evolution of the Ship of Theseus tool itself - meta!

---

## Core Requirements

### 1. PRIMARY ALGORITHM
For each line of code in the current repository state:
1. Use `git blame` to find when the line was last modified
2. Trace that line backwards through git history to find its FIRST appearance
3. Compare the original line to the current line using Levenshtein distance
4. If similarity ≥ 25%, count it as "original"; otherwise it's "completely different"
5. Aggregate statistics across the entire codebase

### 2. LINE MATCHING RULES
- **Similarity Threshold**: 25% minimum (below this = not the same line)
- **Line Movement Detection**: If a line moves within the SAME file and within ±10 lines, it's still "the same line"
- **Rename Following**: Use git's rename detection (`--follow` flag) to track files across renames
- **Similarity Metric**: Use Levenshtein distance formula: `similarity = 1.0 - (distance / max(len(original), len(current)))`

### 3. FILTERING REQUIREMENTS (MUST SKIP)
- Binary files (images, executables, compiled files)
- Generated code (files with `.generated.`, `.gen.`, `.pb.go`, etc.)
- Vendor dependencies (`vendor/`, `node_modules/`, etc.)
- Blank lines
- Comment-only lines (support C-style `//` and `/* */`, Python `#`, etc.)
- Build artifacts (`dist/`, `build/`, `target/`, `__pycache__/`)

### 4. PERFORMANCE REQUIREMENTS
- **Parallelization**: MUST process files in parallel (use worker pool pattern)
- **Git CLI over Library**: Use git CLI commands via `exec.Command` instead of go-git for blame operations (10-100x faster)
- **Progress Indication**: Show real-time progress bar with file count and percentage
- **Memory Management**: Stream results to disk if analyzing >1000 files to avoid memory issues
- **Default Workers**: Use `runtime.NumCPU()` as default worker count

### 5. HISTORICAL TIMELINE
Generate a graph showing code evolution over time:
- Sample every Nth commit (configurable, default: 50)
- Use **heuristic estimation** for "original code %" at each snapshot:
  - Combine exponential time decay with commit churn metrics
  - Formula: `originalPct = 100 * (1 - ageRatio * decayRate * 100) * churnFactor`
  - Get churn from `git show --stat` (additions + deletions)
  - Minimum baseline: 10%
- Display as ASCII graph in terminal

## Technical Specifications

### Dependencies
```go
require (
    github.com/go-git/go-git/v5 v5.11.0
    github.com/agnivade/levenshtein v1.1.1
    github.com/schollz/progressbar/v3 v3.14.1
)
```

### Project Structure
```
ship-of-theseus/
├── main.go                    # CLI entry point
├── go.mod
├── go.sum
├── README.md
├── .gitignore
├── internal/
│   ├── analyzer/
│   │   ├── analyzer.go        # Main orchestration & parallelization
│   │   ├── blame.go           # Git blame via CLI
│   │   ├── history.go         # Line history tracing with rename detection
│   │   ├── snapshots.go       # Historical timeline generation
│   │   └── similarity.go      # Levenshtein similarity calculation
│   ├── filter/
│   │   ├── files.go           # Binary/vendor/generated file detection
│   │   └── comments.go        # Language-specific comment detection
│   ├── visualizer/
│   │   └── graph.go           # Terminal output formatting
│   └── models/
│       └── types.go           # Data structures
```

### Core Data Structures
```go
type CodebaseAnalysis struct {
    TotalLines          int
    OriginalLines       int
    AverageSimilarity   float64
    FileAnalyses        []*FileAnalysis
    HistoricalSnapshots []Snapshot
}

type FileAnalysis struct {
    Path              string
    TotalLines        int
    OriginalLines     int
    AvgSimilarity     float64
    LineHistories     []*LineHistory
}

type LineHistory struct {
    CurrentLine      string
    OriginalLine     string
    CurrentLineNum   int
    OriginalLineNum  int
    FirstCommitHash  string
    FirstCommitDate  time.Time
    LastCommitHash   string
    LastCommitDate   time.Time
    Similarity       float64
}

type Snapshot struct {
    CommitHash   string
    Date         time.Time
    OriginalPct  float64
}
```

### Critical Git CLI Commands to Use

1. **Blame (use this instead of go-git):**
```bash
git -C <repo_path> blame --line-porcelain <file_path>
```
Parse porcelain format for commit hash per line.

2. **File History with Rename Detection:**
```bash
git -C <repo_path> log --follow --pretty=format:%H|%ct -- <file_path>
```

3. **File Content at Commit:**
```bash
git -C <repo_path> show <commit_hash>:<file_path>
```

4. **Commit Statistics:**
```bash
git -C <repo_path> show --stat --pretty=format: <commit_hash>
```

5. **Rename Detection:**
Check patches between commits to detect if file was renamed.

### Algorithm Implementation Details

#### Line Similarity Function
```go
func calculateSimilarity(original, current string) float64 {
    original = strings.TrimSpace(original)
    current = strings.TrimSpace(current)
    
    if original == current { return 1.0 }
    if original == "" || current == "" { return 0.0 }
    
    distance := levenshtein.ComputeDistance(original, current)
    maxLen := max(len(original), len(current))
    
    return 1.0 - (float64(distance) / float64(maxLen))
}
```

#### Finding Similar Line in Range (±10 lines)
```go
func findSimilarLineInRange(lines []string, targetLineNum int, targetLine string) (string, int) {
    start := max(0, targetLineNum-10)
    end := min(len(lines)-1, targetLineNum+10)
    
    bestMatch := ""
    bestSimilarity := 0.0
    bestLineNum := -1
    
    for i := start; i <= end; i++ {
        sim := calculateSimilarity(targetLine, lines[i])
        if sim > bestSimilarity && sim >= 0.25 {
            bestMatch = lines[i]
            bestSimilarity = sim
            bestLineNum = i
        }
    }
    
    return bestMatch, bestLineNum
}
```

#### Comment Detection (Language-Agnostic Regex)
Support these patterns:
- C-style: `//`, `/* */`
- Python/Shell: `#`
- HTML/XML: `<!-- -->`

Use file extension to determine which patterns to apply.

#### Worker Pool Pattern
```go
func processFilesParallel(repoPath string, workItems []WorkItem, numWorkers int) []*FileAnalysis {
    workChan := make(chan WorkItem, len(workItems))
    resultChan := make(chan *FileAnalysis, 100)
    var wg sync.WaitGroup
    
    // Start workers
    for i := 0; i < numWorkers; i++ {
        wg.Add(1)
        go worker(repoPath, workChan, resultChan, &wg)
    }
    
    // Send work
    // Collect results with progress bar
    // Return aggregated results
}
```

## CLI Interface

### Command-Line Flags
```bash
--path string      Path to git repository (default: ".")
--workers int      Number of parallel workers (default: runtime.NumCPU())
--sample int       Sample every Nth commit for history (default: 50)
```

### Output Format
The tool should display:

1. **Header** with ASCII art/emoji
2. **Overall Statistics:**
   - Total lines of code
   - Original lines (count and %)
   - Average similarity %
3. **Interpretation:** Philosophical message based on percentage
4. **Progress Bar:** Visual representation of "original code remaining"
5. **Historical Graph:** ASCII timeline showing evolution
6. **Top 10 Most Transformed Files**
7. **Top 5 Most Stable Files**
8. **Footer** with philosophical reflection

Example output styling:
```
═══════════════════════════════════════════════════════════════════
                    🚢 SHIP OF THESEUS
                 Codebase Evolution Analysis
═══════════════════════════════════════════════════════════════════

📊 OVERALL STATISTICS
   Total Lines of Code:    45,234
   Original Lines:         12,456 (27.5%)
   Average Similarity:     68.3%

💭 INTERPRETATION
   ⚡ This codebase has undergone substantial evolution...

📈 ORIGINAL CODE REMAINING
   [████████████████░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░] 27.5%

📉 EVOLUTION TIMELINE
   (ASCII graph showing percentage over time)

🔥 TOP 10 MOST TRANSFORMED FILES
   1. src/main.go                                      5.2% original
   ...

🏛️  TOP 5 MOST STABLE FILES
   1. config/constants.go                             94.1% original
   ...
```

## Edge Cases & Error Handling

### MUST Handle:
1. **Empty repositories** - graceful error message
2. **Non-git directories** - detect and error early
3. **Binary files** - skip silently
4. **Unreadable files** - skip and continue
5. **Detached HEAD** - still analyze
6. **Very large repos** (>10K files) - streaming to disk
7. **Shallow clones** - work with available history
8. **Corrupted git history** - skip problematic commits
9. **Files deleted and re-added** - treat as new lines
10. **Unicode/non-ASCII content** - handle properly

### Error Messages Should:
- Be user-friendly, not technical stack traces
- Suggest solutions where possible
- Use emoji for visual scanning (❌ for errors, ⚠️ for warnings)

## Testing Strategy

### Test on Real Repositories:
1. **Small repos** (< 100 files): Quick validation
2. **Medium repos** (1K-5K files): Performance testing
3. **Large repos** (>10K files): Stress testing

### Recommended Test Projects:
- `kubernetes/kubernetes` - Very large, actively maintained
- `golang/go` - Mature, stable codebase
- `grafana/grafana` - Good mix of old and new
- `torvalds/linux` - Extremely old codebase
- A small personal project - Easy to verify results manually

### Validation Checks:
- Total line counts should match `git ls-files | xargs wc -l`
- Files in vendor/ should be skipped
- Progress bar should reach 100%
- No panics or crashes on any repo
- Memory usage should stay reasonable (<2GB for large repos)

## Performance Targets
- **Small repo** (<100 files): <30 seconds
- **Medium repo** (1K-5K files): <5 minutes
- **Large repo** (>10K files): <30 minutes
- **Memory**: <2GB peak even for large repos

## Implementation Priority
1. ✅ **Initial setup & data structures** - Commit after completion
2. ✅ **File filtering** - Commit after completion
3. ✅ **Core similarity calculation** - Commit after completion
4. ✅ **Git CLI integration** - Commit after completion
5. ✅ **Line history tracing** - Commit after completion
6. ✅ **Parallelization** - Commit after completion
7. ✅ **Historical timeline** - Commit after completion
8. ✅ **Output formatting** - Commit after completion
9. ✅ **CLI entry point** - Commit after completion
10. ✅ **Testing & bug fixes** - Commit each fix

## Success Criteria
The tool is successful if:
✅ It correctly identifies original vs. modified lines using Levenshtein distance
✅ It follows files through renames using git's `--follow`
✅ It skips binary files, comments, and vendor dependencies
✅ It processes files in parallel with visible progress
✅ It generates both current statistics and historical timeline
✅ It produces beautiful, philosophical output
✅ It handles large repositories without crashing or excessive memory use
✅ It can analyze real-world projects like Kubernetes or Linux
✅ **It has a clean git history showing the development process**

## Additional Considerations

### Future Enhancements (NOT for initial version):
- JSON/CSV export
- HTML report generation
- Git integration (run via `git ship-of-theseus`)
- Comparison between branches
- Contributor-based analysis
- Language-specific metrics

### Code Quality:
- Proper error handling (don't panic)
- Graceful degradation
- Clear variable names
- Comments for complex algorithms
- Modular, testable functions

### Git Best Practices:
- Create `.gitignore` to exclude build artifacts and IDE files
- Include `go.sum` in version control
- Add README.md with:
  - Installation instructions
  - Usage examples
  - Algorithm explanation
  - Performance characteristics
  - Example output screenshots

## Final Notes
This is a philosophical tool as much as a technical one. The output should make developers reflect on the nature of their codebase's evolution. Make the output beautiful, thoughtful, and slightly poetic. The Ship of Theseus paradox should be evident in both the metrics and the presentation.

**Remember**: Commit your work frequently as you build this. The version control history of building this tool is itself a meta-commentary on code evolution!

---

Now implement this complete tool following all specifications above. **Commit each logical group of changes to git as you go.**