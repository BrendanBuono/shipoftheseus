# CLAUDE.md - AI Development Guide

## Overview
This document provides comprehensive guidance for AI assistants (particularly Claude) working on the Ship of Theseus codebase analyzer project. It contains project context, development guidelines, and instructions for maintaining code quality and consistency.

## Project Context

### What is Ship of Theseus?
Ship of Theseus is a CLI tool written in Go that analyzes git repositories to determine how much "original" code remains over time. Named after the ancient philosophical paradox, it traces each line of code from its first appearance to its current state, measuring similarity to answer: "Is this still the same codebase?"

### Philosophical Foundation
The Ship of Theseus paradox asks: If you replace every plank of a ship over time, is it still the same ship? This tool applies that question to codebases:
- If every line of code is eventually modified, is it the same codebase?
- At what point does a project become "different"?
- How do we measure continuity vs. change?

This philosophical depth should be reflected in the tool's output and user experience.

## Project Structure
```
ship-of-theseus/
├── specs/
│   └── core.md              # Complete technical specification
├── main.go                  # CLI entry point
├── go.mod, go.sum          # Go module files
├── README.md               # User-facing documentation
├── CLAUDE.md              # This file - AI development guide
├── .gitignore             # Git ignore patterns
└── internal/
    ├── models/
    │   └── types.go        # Core data structures
    ├── analyzer/
    │   ├── analyzer.go     # Main orchestration
    │   ├── blame.go        # Git CLI integration
    │   ├── history.go      # Line tracing logic
    │   ├── snapshots.go    # Historical analysis
    │   └── similarity.go   # Levenshtein calculations
    ├── filter/
    │   ├── files.go        # File type filtering
    │   └── comments.go     # Comment detection
    └── visualizer/
        └── graph.go        # Terminal output
```

## Core Technical Decisions

### Why Git CLI Over go-git?
**Decision**: Use `os/exec` to call git CLI instead of go-git library for blame operations.
**Reason**: Performance. Git CLI is 10-100x faster for blame operations on large repositories.
**Implementation**: Wrap git commands in helper functions in `internal/analyzer/blame.go`.

### Why Levenshtein Distance?
**Decision**: Use Levenshtein distance for line similarity.
**Reason**: Well-established algorithm that measures edit distance (insertions, deletions, substitutions). Simple, fast, and intuitive.
**Alternative Considered**: Longest Common Subsequence (LCS) - rejected for being too permissive.

### Why 25% Similarity Threshold?
**Decision**: Lines below 25% similarity are considered "completely different."
**Reason**: Balance between strictness and practicality. Testing showed:
- 10% = too permissive (false positives)
- 50% = too strict (simple refactors counted as "new")
- 25% = sweet spot for meaningful similarity

### Why ±10 Line Movement Window?
**Decision**: Allow lines to move within ±10 lines and still be "the same."
**Reason**: Common refactoring operations (adding imports, reordering functions) shouldn't invalidate line identity.

### Why Heuristic Historical Snapshots?
**Decision**: Use heuristic estimation instead of full re-analysis for each historical commit.
**Reason**: Performance. Analyzing every sampled commit would make the tool unusably slow on large repos.
**Formula**: Combines exponential time decay with commit churn metrics.

## Development Workflow

### Before Starting Work
1. **Read the spec**: Always start by reading `specs/core.md`
2. **Understand the philosophy**: This isn't just a metrics tool, it's a reflection on code evolution
3. **Check existing code**: Review what's already implemented before adding new features

### Git Commit Strategy
**CRITICAL**: Commit after each logical unit of work. This project values clean git history.

#### Commit Frequency
Commit after completing:
- Each new file/module
- Each complete feature
- Each bug fix
- Each refactoring
- Updates to documentation

#### Commit Message Format
```
<type>: <short description>

<optional longer description>
<optional implementation notes>
```

**Types**:
- `feat`: New feature
- `fix`: Bug fix
- `refactor`: Code restructuring without behavior change
- `perf`: Performance improvement
- `docs`: Documentation only
- `test`: Adding or fixing tests
- `chore`: Maintenance tasks

**Examples**:
```
feat: Add parallel file processing with worker pool

Implements goroutine-based worker pool to process files concurrently.
Uses buffered channels and WaitGroup for coordination.
Default workers: runtime.NumCPU()
```
```
fix: Handle Unicode properly in similarity calculations

Levenshtein distance was counting Unicode runes incorrectly.
Now uses strings.TrimSpace and proper rune handling.
```
```
perf: Use git CLI instead of go-git for blame operations

Switched from go-git Blame() to exec.Command("git", "blame").
Results in 10-100x performance improvement on large files.
```

### Code Review Checklist
Before committing, verify:
- [ ] Code follows Go conventions (`gofmt`, `golint`)
- [ ] Error handling is comprehensive (no panics)
- [ ] Functions have clear, descriptive names
- [ ] Complex logic has explanatory comments
- [ ] No hardcoded paths or values
- [ ] Memory usage is reasonable
- [ ] Works on example repositories

## Implementation Guidelines

### Error Handling Philosophy
**Principle**: Fail gracefully, never crash.
```go
// ❌ BAD - Panics on error
func analyzeFile(path string) *FileAnalysis {
    content, err := os.ReadFile(path)
    if err != nil {
        panic(err) // Never do this!
    }
    // ...
}

// ✅ GOOD - Returns error, logs context
func analyzeFile(path string) (*FileAnalysis, error) {
    content, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("failed to read %s: %w", path, err)
    }
    // ...
}
```

**Rules**:
1. Always return errors, never panic
2. Wrap errors with context using `fmt.Errorf("context: %w", err)`
3. Log warnings for skipped files, not errors
4. Provide helpful error messages to users

### Performance Considerations

#### Memory Management
```go
// For large repositories (>1000 files), stream to disk
if len(results) > 1000 {
    // Write intermediate results to .ship-of-theseus-results/
    // Clear in-memory results
}
```

#### Parallelization
```go
// Always use worker pools, never spawn unlimited goroutines
numWorkers := runtime.NumCPU()
workChan := make(chan WorkItem, len(items))
resultChan := make(chan *Result, 100) // Buffered!

for i := 0; i < numWorkers; i++ {
    go worker(workChan, resultChan)
}
```

#### Git Operations
```go
// ✅ Fast - Use git CLI
cmd := exec.Command("git", "-C", repoPath, "blame", "--line-porcelain", filePath)

// ❌ Slow - Avoid go-git for blame
blame, _ := git.Blame(repo, filePath) // 10-100x slower
```

### Code Style

#### Naming Conventions
```go
// Types: PascalCase
type LineHistory struct { }
type CodebaseAnalysis struct { }

// Functions: camelCase with descriptive verbs
func analyzeFile() { }
func calculateSimilarity() { }
func traceLineHistory() { }

// Private functions: start with lowercase
func findSimilarLineInRange() { }
func getBlameForFile() { }

// Constants: PascalCase or ALL_CAPS for public
const DefaultSampleRate = 50
const MinimumSimilarityThreshold = 0.25
```

#### Function Organization
```go
// Public API functions first
func AnalyzeRepository(...) (*CodebaseAnalysis, error) { }

// Then private helpers
func processFilesParallel(...) []*FileAnalysis { }
func analyzeFile(...) *FileAnalysis { }
func traceLineHistory(...) *LineHistory { }
```

#### Comments
```go
// Package-level comments explain the purpose
// Package analyzer implements the core codebase analysis logic,
// tracing lines through git history to determine originality.
package analyzer

// Function comments explain what, not how (code shows how)
// calculateSimilarity compares two strings using Levenshtein distance
// and returns a value between 0.0 (completely different) and 1.0 (identical).
func calculateSimilarity(original, current string) float64 {
    // Complex algorithms get inline explanation
    // Levenshtein distance counts the minimum edits needed to transform
    // one string into another (insertions, deletions, substitutions)
    distance := levenshtein.ComputeDistance(original, current)
    // ...
}
```

### Testing Strategy

#### Manual Testing
Always test on these repository types:
1. **Tiny repo** (1-10 files) - Your own test repo
2. **Small repo** (10-100 files) - Small OSS project
3. **Medium repo** (100-1K files) - Mid-size OSS project
4. **Large repo** (1K-10K files) - `grafana/grafana`, `golang/go`
5. **Huge repo** (10K+ files) - `kubernetes/kubernetes`

#### Validation Checks
```bash
# Line count should roughly match
git ls-files | grep -v vendor | xargs wc -l

# No panics or crashes
./ship-of-theseus --path /path/to/large/repo

# Memory stays reasonable
# Monitor with: ps aux | grep ship-of-theseus
# Should stay under 2GB even for huge repos

# Progress bar reaches 100%
# Visual verification during run
```

#### Edge Case Testing
Test these scenarios:
- Empty repository
- Single-file repository
- Repository with binary files only
- Repository with no commit history
- Shallow clone (--depth 1)
- Repository with merge conflicts
- Repository with deleted and re-added files
- Non-UTF8 file encodings

## Common Pitfalls & Solutions

### Pitfall 1: Ignoring Renames
**Problem**: Not using `--follow` when tracing file history causes lost lineage.
**Solution**: Always use `git log --follow` for file history.

### Pitfall 2: Memory Leaks with Large Repos
**Problem**: Keeping all results in memory crashes on huge repositories.
**Solution**: Stream to disk after 1000 files, aggregate at the end.

### Pitfall 3: Slow Blame Operations
**Problem**: Using go-git's Blame() is very slow.
**Solution**: Use `git blame --line-porcelain` via exec.Command.

### Pitfall 4: Not Handling Binary Files
**Problem**: Trying to analyze binary files causes errors.
**Solution**: Check file extensions and skip early.

### Pitfall 5: Incorrect Similarity Calculation
**Problem**: Not trimming whitespace skews similarity scores.
**Solution**: Always `strings.TrimSpace()` before comparing.

### Pitfall 6: Progress Bar Not Reaching 100%
**Problem**: Lost work items or incorrect counting.
**Solution**: Carefully track sent vs. processed items.

## Extending the Project

### Adding New Filters
Location: `internal/filter/`
```go
// files.go - Add new file patterns
var skipDirectories = map[string]bool{
    "vendor":       true,
    "node_modules": true,
    "your_new_dir": true, // Add here
}

// comments.go - Add new language support
var commentPatterns = map[string][]*regexp.Regexp{
    "rust": {
        regexp.MustCompile(`^\s*//`),
        regexp.MustCompile(`^\s*/\*`),
    },
}
```

### Adding New Output Formats
Location: `internal/visualizer/graph.go`
```go
// Add JSON export
func ExportJSON(analysis *CodebaseAnalysis) ([]byte, error) {
    return json.MarshalIndent(analysis, "", "  ")
}

// Add HTML report
func GenerateHTMLReport(analysis *CodebaseAnalysis) string {
    // Template-based HTML generation
}
```

### Adding New Metrics
Location: `internal/analyzer/analyzer.go`
```go
// Example: Track churn rate per directory
type DirectoryStats struct {
    Path         string
    ChurnRate    float64
    OriginalPct  float64
}

func calculateDirectoryStats(analysis *CodebaseAnalysis) []DirectoryStats {
    // Group files by directory
    // Calculate per-directory metrics
}
```

## Debugging Tips

### Enable Verbose Logging
Add a `--verbose` flag for debugging:
```go
if verbose {
    log.Printf("Processing file: %s (line count: %d)", path, lineCount)
    log.Printf("Found blame info: %v", blameInfo)
    log.Printf("Similarity: %.2f%%", similarity*100)
}
```

### Profile Memory Usage
```go
import _ "net/http/pprof"

// In main():
go func() {
    log.Println(http.ListenAndServe("localhost:6060", nil))
}()

// Then: go tool pprof http://localhost:6060/heap
```

### Profile CPU Usage
```go
import "runtime/pprof"

f, _ := os.Create("cpu.prof")
pprof.StartCPUProfile(f)
defer pprof.StopCPUProfile()

// Then: go tool pprof cpu.prof
```

### Test Single File
Add a debug mode to analyze just one file:
```go
if debugFile != "" {
    analysis := analyzeFile(repoPath, debugFile, content)
    fmt.Printf("%+v\n", analysis)
    os.Exit(0)
}
```

## Documentation Standards

### README.md Structure
1. **Hero section** - One-line description, badges
2. **Installation** - Step-by-step setup
3. **Quick start** - Basic usage example
4. **How it works** - Algorithm explanation
5. **Options** - CLI flag documentation
6. **Examples** - Real-world usage scenarios
7. **Performance** - Benchmarks and targets
8. **Philosophy** - Ship of Theseus explanation
9. **Contributing** - How to contribute
10. **License** - MIT or similar

### Inline Documentation
```go
// Package-level godoc
// Package analyzer traces lines through git history to determine
// what percentage of a codebase remains "original" over time.
//
// The core algorithm uses git blame to find when each line was
// last modified, then traces backwards through commits to find
// the first appearance. Levenshtein distance measures similarity.
package analyzer

// Exported functions need godoc
// AnalyzeRepository performs a complete Ship of Theseus analysis
// on the given repository path. It returns aggregated statistics
// about code originality across the entire codebase.
//
// The analysis runs in parallel using numWorkers goroutines.
// Historical snapshots are generated by sampling every sampleRate commits.
func AnalyzeRepository(repoPath string, numWorkers, sampleRate int) (*CodebaseAnalysis, error) {
```

## FAQs for AI Assistants

### Q: Should I use go-git or git CLI?
**A**: Use go-git for repo-level operations (getting commits, trees). Use git CLI (exec.Command) for blame, log, and show operations. It's hybrid by design.

### Q: How do I handle files that moved between directories?
**A**: Use `git log --follow` which tracks renames. The `traceFileRenames()` function handles this.

### Q: What if similarity calculation is too slow?
**A**: The `agnivade/levenshtein` package is already optimized. If still slow, consider:
1. Caching similarity results
2. Short-circuit on exact matches
3. Parallel similarity calculations

### Q: How do I add support for a new language's comments?
**A**: Edit `internal/filter/comments.go` and add patterns to `commentPatterns` map.

### Q: Should I add features not in the spec?
**A**: No. Stick to `specs/core.md`. Mark future ideas with `// TODO:` comments for later.

### Q: How do I test rename detection?
**A**: Create a test repo, add a file, commit, `git mv` the file, commit, modify the file, commit. Then analyze.

### Q: What if a user reports incorrect percentages?
**A**: Debug checklist:
1. Verify comment filtering is working
2. Check if binary files are being skipped
3. Verify similarity threshold (25%) is correct
4. Test on a small repo where you can manually verify

### Q: How do I make the tool faster?
**A**: Priority order:
1. Increase parallelism (more workers)
2. Use git CLI for expensive operations
3. Cache git operations (commits, file contents)
4. Stream results to disk for huge repos

## Release Checklist

Before releasing a version:
- [ ] All tests pass on small, medium, and large repos
- [ ] Memory usage is reasonable (<2GB for large repos)
- [ ] No panics or crashes
- [ ] Progress bar works correctly
- [ ] Output formatting is beautiful
- [ ] README is complete and accurate
- [ ] CLAUDE.md is updated with any new patterns
- [ ] Git history is clean and descriptive
- [ ] Code passes `gofmt` and `golint`
- [ ] Version number updated in main.go
- [ ] Git tag created: `git tag v1.0.0`

## Philosophy Reminders

This tool is **both technical and philosophical**:
- The output should inspire reflection on code evolution
- Use poetic language in output ("Like a well-maintained ship...")
- Include emoji thoughtfully (🚢 ⚡ 🔥 🏛️)
- The tool itself is a Ship of Theseus - maintain clean git history to demonstrate this
- Balance precision with performance - perfect accuracy isn't the goal, insightful metrics are

## Resources

- **Spec**: `specs/core.md` - Complete technical specification
- **Git docs**: https://git-scm.com/docs - Git CLI reference
- **Go docs**: https://golang.org/doc/ - Go language reference
- **Levenshtein**: https://en.wikipedia.org/wiki/Levenshtein_distance - Algorithm explanation

---

**Remember**: You're building a tool that makes developers think deeply about their code's evolution. Make it beautiful, make it fast, make it thoughtful.

*"The ship wherein Theseus and the youth of Athens returned had thirty oars, and was preserved by the Athenians down even to the time of Demetrius Phalereus, for they took away the old planks as they decayed, putting in new and stronger timber in their place..."* - Plutarch