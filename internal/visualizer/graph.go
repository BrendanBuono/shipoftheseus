// Package visualizer provides beautiful terminal output for analysis results.
// It creates ASCII art, graphs, and philosophical commentary on code evolution.
package visualizer

import (
	"fmt"
	"ship-of-theseus/internal/models"
	"sort"
	"strings"
)

// Display prints the complete analysis results to the terminal with beautiful formatting.
func Display(analysis *models.CodebaseAnalysis) {
	printHeader()
	printOverallStats(analysis)
	printInterpretation(analysis)
	printProgressBar(analysis)

	if len(analysis.HistoricalSnapshots) > 0 {
		printTimeline(analysis.HistoricalSnapshots)
	}

	printTopTransformed(analysis)
	printTopStable(analysis)
	printFooter(analysis)
}

// printHeader displays the ASCII art title banner.
func printHeader() {
	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════════════════")
	fmt.Println("                    🚢 SHIP OF THESEUS")
	fmt.Println("                 Codebase Evolution Analysis")
	fmt.Println("═══════════════════════════════════════════════════════════════════")
	fmt.Println()
}

// printOverallStats displays repository-wide metrics.
func printOverallStats(analysis *models.CodebaseAnalysis) {
	originalPct := 0.0
	if analysis.TotalLines > 0 {
		originalPct = float64(analysis.OriginalLines) / float64(analysis.TotalLines) * 100.0
	}

	fmt.Println("📊 OVERALL STATISTICS")
	fmt.Printf("   Total Lines of Code:    %s\n", formatNumber(analysis.TotalLines))
	fmt.Printf("   Original Lines:         %s (%.1f%%)\n",
		formatNumber(analysis.OriginalLines), originalPct)
	fmt.Printf("   Average Similarity:     %.1f%%\n", analysis.AverageSimilarity*100)
	fmt.Println()
}

// printInterpretation provides philosophical context based on the originality percentage.
func printInterpretation(analysis *models.CodebaseAnalysis) {
	originalPct := 0.0
	if analysis.TotalLines > 0 {
		originalPct = float64(analysis.OriginalLines) / float64(analysis.TotalLines) * 100.0
	}

	fmt.Println("💭 INTERPRETATION")

	var emoji, message string
	switch {
	case originalPct >= 80:
		emoji = "🏛️"
		message = "This codebase is remarkably stable, like a well-preserved ancient monument.\n   Most code remains close to its original form."
	case originalPct >= 60:
		emoji = "⚓"
		message = "This codebase maintains strong continuity with its origins.\n   Core structures persist through evolution."
	case originalPct >= 40:
		emoji = "🌊"
		message = "This codebase has undergone substantial evolution.\n   Like Theseus's ship, many planks have been replaced."
	case originalPct >= 20:
		emoji = "⚡"
		message = "This codebase has been heavily transformed.\n   Few traces of the original code remain."
	default:
		emoji = "🔥"
		message = "This codebase has been completely reimagined.\n   It bears little resemblance to its origins."
	}

	fmt.Printf("   %s %s\n", emoji, message)
	fmt.Println()
}

// printProgressBar shows a visual representation of originality percentage.
func printProgressBar(analysis *models.CodebaseAnalysis) {
	originalPct := 0.0
	if analysis.TotalLines > 0 {
		originalPct = float64(analysis.OriginalLines) / float64(analysis.TotalLines) * 100.0
	}

	fmt.Println("📈 ORIGINAL CODE REMAINING")

	// Create a 60-character progress bar
	barWidth := 60
	filledWidth := int(originalPct / 100.0 * float64(barWidth))

	bar := "["
	for i := 0; i < barWidth; i++ {
		if i < filledWidth {
			bar += "█"
		} else {
			bar += "░"
		}
	}
	bar += "]"

	fmt.Printf("   %s %.1f%%\n", bar, originalPct)
	fmt.Println()
}

// printTimeline displays an ASCII graph of code evolution over time.
func printTimeline(snapshots []models.Snapshot) {
	if len(snapshots) < 2 {
		return
	}

	fmt.Println("📉 EVOLUTION TIMELINE")
	fmt.Println()

	// Graph parameters
	height := 10
	width := 60

	// Find min/max for scaling
	minPct, maxPct := 100.0, 0.0
	for _, s := range snapshots {
		if s.OriginalPct < minPct {
			minPct = s.OriginalPct
		}
		if s.OriginalPct > maxPct {
			maxPct = s.OriginalPct
		}
	}

	// Add some padding
	pctRange := maxPct - minPct
	if pctRange < 10 {
		pctRange = 10
	}
	minPct -= pctRange * 0.1
	maxPct += pctRange * 0.1
	if minPct < 0 {
		minPct = 0
	}
	if maxPct > 100 {
		maxPct = 100
	}

	// Draw graph from top to bottom
	for y := height; y >= 0; y-- {
		// Calculate the percentage this row represents
		rowPct := minPct + (maxPct-minPct)*float64(y)/float64(height)

		// Y-axis label
		fmt.Printf("   %5.0f%% │", rowPct)

		// Plot points
		for x := 0; x < width; x++ {
			// Map x position to snapshot index
			idx := int(float64(x) / float64(width) * float64(len(snapshots)-1))
			snapPct := snapshots[idx].OriginalPct

			// Check if this snapshot's value is close to this row
			if snapPct >= rowPct-((maxPct-minPct)/float64(height)/2) &&
				snapPct <= rowPct+((maxPct-minPct)/float64(height)/2) {
				fmt.Print("●")
			} else if snapPct > rowPct {
				// Value is above this row
				fmt.Print(" ")
			} else {
				// Value is below this row
				fmt.Print(" ")
			}
		}
		fmt.Println()
	}

	// X-axis
	fmt.Print("          └")
	fmt.Print(strings.Repeat("─", width))
	fmt.Println("▶")
	fmt.Printf("           %s%s%s\n",
		snapshots[0].Date.Format("2006"),
		strings.Repeat(" ", width-12),
		snapshots[len(snapshots)-1].Date.Format("2006"))
	fmt.Println()
}

// printTopTransformed shows the files that have changed the most.
func printTopTransformed(analysis *models.CodebaseAnalysis) {
	// Sort files by original percentage (ascending)
	sorted := make([]*models.FileAnalysis, len(analysis.FileAnalyses))
	copy(sorted, analysis.FileAnalyses)

	sort.Slice(sorted, func(i, j int) bool {
		pctI := float64(sorted[i].OriginalLines) / float64(sorted[i].TotalLines)
		pctJ := float64(sorted[j].OriginalLines) / float64(sorted[j].TotalLines)
		return pctI < pctJ
	})

	fmt.Println("🔥 TOP 10 MOST TRANSFORMED FILES")
	count := 10
	if len(sorted) < count {
		count = len(sorted)
	}

	for i := 0; i < count; i++ {
		file := sorted[i]
		pct := float64(file.OriginalLines) / float64(file.TotalLines) * 100.0
		fmt.Printf("   %2d. %-50s %5.1f%% original\n", i+1, truncatePath(file.Path, 50), pct)
	}
	fmt.Println()
}

// printTopStable shows the files that have changed the least.
func printTopStable(analysis *models.CodebaseAnalysis) {
	// Sort files by original percentage (descending)
	sorted := make([]*models.FileAnalysis, len(analysis.FileAnalyses))
	copy(sorted, analysis.FileAnalyses)

	sort.Slice(sorted, func(i, j int) bool {
		pctI := float64(sorted[i].OriginalLines) / float64(sorted[i].TotalLines)
		pctJ := float64(sorted[j].OriginalLines) / float64(sorted[j].TotalLines)
		return pctI > pctJ
	})

	fmt.Println("🏛️  TOP 5 MOST STABLE FILES")
	count := 5
	if len(sorted) < count {
		count = len(sorted)
	}

	for i := 0; i < count; i++ {
		file := sorted[i]
		pct := float64(file.OriginalLines) / float64(file.TotalLines) * 100.0
		fmt.Printf("   %d. %-50s %5.1f%% original\n", i+1, truncatePath(file.Path, 50), pct)
	}
	fmt.Println()
}

// printFooter displays a philosophical closing message.
func printFooter(analysis *models.CodebaseAnalysis) {
	fmt.Println("═══════════════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Println("  \"The ship wherein Theseus and the youth of Athens returned had")
	fmt.Println("   thirty oars, and was preserved by the Athenians... for they took")
	fmt.Println("   away the old planks as they decayed, putting in new and stronger")
	fmt.Println("   timber in their place...\"")
	fmt.Println()
	fmt.Println("                                                    — Plutarch")
	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════════════════")
	fmt.Println()
}

// formatNumber adds comma separators to large numbers for readability.
func formatNumber(n int) string {
	str := fmt.Sprintf("%d", n)
	if len(str) <= 3 {
		return str
	}

	var result string
	for i, digit := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result += ","
		}
		result += string(digit)
	}
	return result
}

// truncatePath shortens a file path to fit within maxLen characters.
func truncatePath(path string, maxLen int) string {
	if len(path) <= maxLen {
		return path
	}

	// Truncate from the left, keeping the filename visible
	return "..." + path[len(path)-maxLen+3:]
}
