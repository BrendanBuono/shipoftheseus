package analyzer

import (
	"fmt"
	"math"
	"ship-of-theseus/internal/models"
)

// GenerateHistoricalSnapshots creates a timeline showing how code originality changed over time.
// Uses heuristic estimation rather than full re-analysis of each commit for performance.
//
// Parameters:
//   - repoPath: Path to git repository
//   - sampleRate: Analyze every Nth commit (e.g., 50 = every 50th commit)
//   - currentOriginalPct: The current percentage of original code (from main analysis)
//
// Algorithm:
//  1. Get all commits in repository
//  2. Sample every Nth commit
//  3. For each sample, estimate originality using:
//     - Time decay: older commits had more "original" code
//     - Churn factor: commits with high churn reduce originality more
//  4. Apply formula: originalPct = 100 * (1 - ageRatio * decayRate) * churnFactor
//  5. Minimum baseline: 10% (code rarely drops below this)
//
// PHILOSOPHICAL NOTE: Originality Can Increase Over Time
// ======================================================
// This heuristic allows originality percentages to INCREASE at certain points in the timeline.
// This is intentional and philosophically meaningful. When code refactors toward simplicity
// or reverts complex changes, it becomes more similar to its original form - the Ship of
// Theseus getting its old planks back. This measures "snapshot similarity to origin", not
// "accumulated irreversible change". A codebase that simplifies after experimentation is
// becoming MORE original, and that's worth celebrating.
func GenerateHistoricalSnapshots(repoPath string, sampleRate int, currentOriginalPct float64) ([]models.Snapshot, error) {
	// Get all commits
	allCommits, err := GetAllCommits(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get commits: %w", err)
	}

	if len(allCommits) == 0 {
		return nil, fmt.Errorf("no commits found in repository")
	}

	// Sample commits
	var sampledCommits []CommitInfo
	for i := len(allCommits) - 1; i >= 0; i -= sampleRate {
		sampledCommits = append(sampledCommits, allCommits[i])
	}

	// Always include the most recent commit
	if len(sampledCommits) == 0 || sampledCommits[len(sampledCommits)-1].Hash != allCommits[0].Hash {
		sampledCommits = append(sampledCommits, allCommits[0])
	}

	// Generate snapshots using heuristic estimation
	snapshots := make([]models.Snapshot, 0, len(sampledCommits))

	for i, commit := range sampledCommits {
		// Calculate age ratio (0.0 = oldest, 1.0 = newest)
		ageRatio := float64(i) / float64(len(sampledCommits)-1)
		if len(sampledCommits) == 1 {
			ageRatio = 1.0
		}

		// Get commit churn stats
		stats, err := GetCommitStats(repoPath, commit.Hash)
		if err != nil {
			// If we can't get stats, use neutral churn factor
			stats = map[string]int{"additions": 0, "deletions": 0}
		}

		// Calculate churn factor (higher churn = more change = less originality)
		totalChurn := stats["additions"] + stats["deletions"]
		churnFactor := calculateChurnFactor(totalChurn)

		// Estimate original percentage using time decay and churn
		// Formula: Start at 100%, decay based on age, adjust for churn
		// The most recent commit should match our current analysis
		var originalPct float64
		if i == len(sampledCommits)-1 {
			// For the most recent commit, use the actual calculated percentage
			originalPct = currentOriginalPct
		} else {
			// For historical commits, estimate based on age and churn
			// Base assumption: older code was more "original"
			decayRate := 0.6 // How fast originality decays (0.0 = no decay, 1.0 = full decay)
			baseOriginality := 100.0

			// Apply exponential decay with age
			ageDecay := 1.0 - (ageRatio * decayRate)

			// Apply churn factor
			originalPct = baseOriginality * ageDecay * churnFactor

			// Clamp to minimum baseline (code rarely drops below 10%)
			if originalPct < 10.0 {
				originalPct = 10.0
			}

			// Ensure we don't exceed current originality (can't be more original in past
			// than present if we're measuring from present)
			if originalPct < currentOriginalPct {
				originalPct = currentOriginalPct + (ageDecay * (100.0 - currentOriginalPct) * 0.3)
			}
		}

		snapshots = append(snapshots, models.Snapshot{
			CommitHash:  commit.Hash,
			Date:        commit.Date,
			OriginalPct: originalPct,
		})
	}

	return snapshots, nil
}

// calculateChurnFactor converts commit churn (lines changed) into a factor between 0.0 and 1.0.
// Higher churn = lower factor = more impact on originality.
//
// Churn levels:
//   - 0-50 lines: 1.0 (minimal impact)
//   - 50-200 lines: 0.9-0.7 (moderate impact)
//   - 200-1000 lines: 0.7-0.5 (significant impact)
//   - 1000+ lines: 0.5-0.3 (major refactoring)
func calculateChurnFactor(totalChurn int) float64 {
	if totalChurn <= 50 {
		return 1.0
	} else if totalChurn <= 200 {
		// Linear interpolation between 1.0 and 0.7
		ratio := float64(totalChurn-50) / 150.0
		return 1.0 - (ratio * 0.3)
	} else if totalChurn <= 1000 {
		// Linear interpolation between 0.7 and 0.5
		ratio := float64(totalChurn-200) / 800.0
		return 0.7 - (ratio * 0.2)
	} else {
		// Asymptotic approach to 0.3 for very large changes
		// Use exponential decay: 0.5 * exp(-x/2000) + 0.3
		excess := float64(totalChurn - 1000)
		return 0.5*math.Exp(-excess/2000.0) + 0.3
	}
}

// AddSnapshotsToAnalysis updates an analysis with historical snapshots.
func AddSnapshotsToAnalysis(analysis *models.CodebaseAnalysis, repoPath string, sampleRate int) error {
	currentPct := 0.0
	if analysis.TotalLines > 0 {
		currentPct = float64(analysis.OriginalLines) / float64(analysis.TotalLines) * 100.0
	}

	snapshots, err := GenerateHistoricalSnapshots(repoPath, sampleRate, currentPct)
	if err != nil {
		return err
	}

	analysis.HistoricalSnapshots = snapshots
	return nil
}
