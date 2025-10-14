# 🚢 Ship of Theseus

**A philosophical Git codebase evolution analyzer**

> _"The ship wherein Theseus and the youth of Athens returned had thirty oars, and was preserved by the Athenians... for they took away the old planks as they decayed, putting in new and stronger timber in their place..."_ — Plutarch

Ship of Theseus traces each line of code from its first appearance to its current state, measuring similarity to answer the ancient philosophical question: **Is this still the same codebase?**

## The Paradox

If you replace every plank of a ship over time, is it still the same ship? Similarly, if every line of code is eventually modified, is it still the same codebase? This tool measures how much "original" code remains by analyzing git history.

## Features

- ⚡ **Fast**: Uses git CLI (10-100x faster than libraries) and parallel processing
- 📊 **Comprehensive**: Traces every line through git history with rename detection
- 🎨 **Beautiful**: ASCII art, graphs, and philosophical commentary
- 🔍 **Smart**: Uses Levenshtein distance to measure code similarity
- 🏛️ **Historical**: Generates timeline showing code evolution over time
- 🎯 **Accurate**: Filters comments, blanks, generated code, and vendor dependencies

## Installation

### From Source

Requires Go 1.21 or later:

```bash
git clone https://github.com/yourusername/ship-of-theseus.git
cd ship-of-theseus
go build -o ship-of-theseus .
sudo mv ship-of-theseus /usr/local/bin/
```

### Using Go Install

```bash
go install github.com/yourusername/ship-of-theseus@latest
```

## Quick Start

```bash
# Analyze current directory
ship-of-theseus

# Analyze specific repository
ship-of-theseus --path /path/to/repo

# Use more workers for faster analysis
ship-of-theseus --workers 16

# Coarser sampling for speed
ship-of-theseus --sample 100
```

## How It Works

### The Algorithm

For each line of code in the current repository:

1. **Get Tracked Files**: Use `git ls-files` to get all tracked files (respects .gitignore)
2. **Read from Git**: Get file content from `git show HEAD:file` (not working directory)
3. **Get First Commit**: Find the file's first commit using `git log --follow`
4. **Compare First to Current**: Compare line from first commit to current line
5. **Find Similar Lines**: Look for similar lines within ±10 lines of position
6. **Measure Similarity**: Use Levenshtein distance to calculate percentage similarity
7. **Determine Originality**: Lines with ≥25% similarity are "original"

### Similarity Threshold

**Why 25%?** Testing showed this is the sweet spot:
- **10%**: Too permissive (false positives)
- **25%**: Balanced (meaningful similarity)
- **50%**: Too strict (simple refactors counted as "new")

### Line Movement

Lines can move within **±10 lines** and still be considered "the same line". This handles common refactoring like:
- Adding imports
- Reordering functions
- Extracting methods

### What Gets Skipped

**Automatically (via `git ls-files`):**
- All files in `.gitignore` (logs, `.env`, build artifacts, etc.)
- Untracked files
- Ignored directories (`.git/`, etc.)

**By file filtering:**
- Binary files (images, executables, archives)
- Generated code (`.pb.go`, `.gen.`, `_generated.`)
- Vendor dependencies if committed (`vendor/`, `node_modules/`)

**By line filtering:**
- Blank lines
- Comment-only lines (supports 30+ languages)

## Output Explained

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
   ⚡ This codebase has undergone substantial evolution.
   Like Theseus's ship, many planks have been replaced.

📈 ORIGINAL CODE REMAINING
   [████████████████░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░] 27.5%
```

### Interpretation Guide

- **🏛️ 80-100%**: Remarkably stable, well-preserved
- **⚓ 60-80%**: Strong continuity, core structures persist
- **🌊 40-60%**: Substantial evolution, many changes
- **⚡ 20-40%**: Heavily transformed, few original traces
- **🔥 0-20%**: Completely reimagined, barely recognizable

## Command-Line Options

```
--path string     Path to git repository (default: ".")
--workers int     Number of parallel workers (default: NumCPU)
--sample int      Sample every Nth commit for timeline (default: 50)
--version         Show version information
```

### Performance Tuning

**Workers**: More workers = faster analysis (diminishing returns beyond NumCPU)
```bash
ship-of-theseus --workers 16  # Use 16 parallel workers
```

**Sampling**: Higher sampling = faster but less accurate timeline
```bash
ship-of-theseus --sample 100  # Sample every 100th commit instead of 50th
```

## Performance

Tested on real repositories:

| Repository | Tracked Files | Analysis Time | Notes |
|---|---|---|---|
| ship-of-theseus | 16 | <5s | Brand new project, 100% original |
| bet-data (8mo old) | 305 | ~60s | 6-100% originality range |

**Estimated performance** (untested):
- Small repos (<100 files): <30 seconds
- Medium repos (500-1K files): 1-5 minutes
- Large repos (5K+ files): 10-30+ minutes

Performance scales with:
- Number of files
- File size (lines per file)
- Commit history depth
- Number of parallel workers

## Technical Details

### Architecture

```
ship-of-theseus/
├── main.go                      # CLI entry point
├── internal/
│   ├── models/types.go         # Core data structures
│   ├── analyzer/
│   │   ├── analyzer.go         # Main orchestration (parallel processing)
│   │   ├── blame.go            # Git CLI wrapper (10-100x faster than libraries)
│   │   ├── history.go          # Line history tracing with rename detection
│   │   ├── snapshots.go        # Historical timeline generation
│   │   └── similarity.go       # Levenshtein distance calculations
│   ├── filter/
│   │   ├── files.go            # Binary/vendor/generated file detection
│   │   └── comments.go         # Language-specific comment detection
│   └── visualizer/
│       └── graph.go            # Terminal output formatting
```

### Why Git CLI vs Libraries?

**Performance**: Git CLI is 10-100x faster for `blame` operations:
- `go-git` Blame(): ~30s for large file
- `git blame` CLI: ~0.3s for same file

We use a hybrid approach:
- **git CLI**: For blame, log, show (performance-critical)
- **go-git**: For repository metadata (when needed)

### Levenshtein Distance

Measures edit distance between strings (insertions, deletions, substitutions):

```
similarity = 1.0 - (distance / max(len(original), len(current)))
```

Example:
```go
CalculateSimilarity("func add(a, b)", "func add(x, y)") → ~0.85 (85%)
CalculateSimilarity("hello world", "goodbye") → ~0.18 (18%)
```

### Historical Timeline

Uses **heuristic estimation** rather than full re-analysis:
- **Time decay**: Older commits had more "original" code
- **Churn factor**: Commits with high churn reduce originality
- **Formula**: `originalPct = 100 * (1 - ageRatio * decayRate) * churnFactor`

This is much faster than re-analyzing every commit, with acceptable accuracy trade-off.

## Contributing

Contributions welcome! Areas for improvement:

- **JSON/CSV export** for data analysis
- **HTML report generation** with interactive graphs
- **Comparison between branches** or tags
- **Contributor-based analysis** (who rewrites the most?)
- **Language-specific metrics** (Go vs JavaScript originality)

## Philosophy

This tool isn't just about metrics—it's about understanding the nature of software evolution. Code is living; it grows, changes, heals, and sometimes dies. Like Theseus's ship, the question isn't whether change happens, but what identity means in the face of continuous transformation.

**Questions to ponder:**
- Is a completely rewritten codebase "new" if it serves the same purpose?
- Does a stable codebase indicate quality or stagnation?
- If AI rewrites all your code, is it still "your" project?

## License

MIT License - see LICENSE file for details

## Acknowledgments

- Inspired by the ancient Ship of Theseus paradox
- Built with [levenshtein](https://github.com/agnivade/levenshtein) for similarity calculations
- Uses Git's powerful history tracking and rename detection

---

*"Is it the same code? Yes and no. It is neither—and both."*
