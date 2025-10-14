package analyzer

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// BlameInfo represents blame information for a single line in a file.
type BlameInfo struct {
	CommitHash string    // Full commit hash
	LineNum    int       // Line number (1-indexed)
	CommitDate time.Time // Commit date
}

// GetBlame runs git blame on a file and returns blame info for each line.
// Uses --line-porcelain format for detailed, machine-readable output.
// This is 10-100x faster than using go-git's Blame() function.
func GetBlame(repoPath, filePath string) ([]BlameInfo, error) {
	// Run: git -C <repo> blame --line-porcelain <file>
	cmd := exec.Command("git", "-C", repoPath, "blame", "--line-porcelain", filePath)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git blame failed for %s: %w", filePath, err)
	}

	return parseBlameOutput(output)
}

// parseBlameOutput parses the porcelain format output from git blame.
// Porcelain format provides commit hash, metadata, and line content in a structured way.
func parseBlameOutput(output []byte) ([]BlameInfo, error) {
	var result []BlameInfo
	scanner := bufio.NewScanner(bytes.NewReader(output))

	var currentHash string
	var currentTime time.Time
	var currentLine int

	for scanner.Scan() {
		line := scanner.Text()

		// Each blame block starts with: <hash> <original-line> <final-line> [<num-lines>]
		if len(line) > 40 && line[40] == ' ' {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				currentHash = parts[0]
				// parts[2] is the final line number
				if lineNum, err := strconv.Atoi(parts[2]); err == nil {
					currentLine = lineNum
				}
			}
			continue
		}

		// committer-time is the Unix timestamp
		if strings.HasPrefix(line, "committer-time ") {
			timestampStr := strings.TrimPrefix(line, "committer-time ")
			if timestamp, err := strconv.ParseInt(timestampStr, 10, 64); err == nil {
				currentTime = time.Unix(timestamp, 0)
			}
			continue
		}

		// The actual line content starts with "\t"
		if strings.HasPrefix(line, "\t") {
			if currentHash != "" && currentLine > 0 {
				result = append(result, BlameInfo{
					CommitHash: currentHash,
					LineNum:    currentLine,
					CommitDate: currentTime,
				})
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error parsing blame output: %w", err)
	}

	return result, nil
}

// GetFileHistory retrieves the commit history for a file, following renames.
// Returns a list of commit hashes in reverse chronological order (newest first).
func GetFileHistory(repoPath, filePath string) ([]string, error) {
	// Run: git -C <repo> log --follow --pretty=format:%H -- <file>
	cmd := exec.Command("git", "-C", repoPath, "log", "--follow", "--pretty=format:%H", "--", filePath)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git log failed for %s: %w", filePath, err)
	}

	// Split output by newlines to get individual commit hashes
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return nil, fmt.Errorf("no history found for %s", filePath)
	}

	return lines, nil
}

// GetFileAtCommit retrieves the contents of a file at a specific commit.
// Returns the file contents as a string.
func GetFileAtCommit(repoPath, commitHash, filePath string) (string, error) {
	// Run: git -C <repo> show <commit>:<file>
	cmd := exec.Command("git", "-C", repoPath, "show", commitHash+":"+filePath)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git show failed for %s at %s: %w", filePath, commitHash, err)
	}

	return string(output), nil
}

// GetCommitStats retrieves statistics about a commit (additions, deletions).
// Returns a map with keys: "additions" and "deletions" as integers.
func GetCommitStats(repoPath, commitHash string) (map[string]int, error) {
	// Run: git -C <repo> show --stat --pretty=format: <commit>
	cmd := exec.Command("git", "-C", repoPath, "show", "--stat", "--pretty=format:", commitHash)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git show --stat failed for %s: %w", commitHash, err)
	}

	return parseCommitStats(output)
}

// parseCommitStats extracts addition/deletion counts from git show --stat output.
// Example line: " 3 files changed, 45 insertions(+), 12 deletions(-)"
func parseCommitStats(output []byte) (map[string]int, error) {
	stats := map[string]int{
		"additions": 0,
		"deletions": 0,
	}

	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()

		// Look for the summary line with insertions/deletions
		if strings.Contains(line, "insertion") || strings.Contains(line, "deletion") {
			// Parse additions
			if idx := strings.Index(line, "insertion"); idx != -1 {
				// Extract number before "insertion"
				parts := strings.Fields(line[:idx])
				if len(parts) > 0 {
					if num, err := strconv.Atoi(parts[len(parts)-1]); err == nil {
						stats["additions"] = num
					}
				}
			}

			// Parse deletions
			if idx := strings.Index(line, "deletion"); idx != -1 {
				// Extract number before "deletion"
				parts := strings.Fields(line[:idx])
				if len(parts) > 0 {
					if num, err := strconv.Atoi(parts[len(parts)-1]); err == nil {
						stats["deletions"] = num
					}
				}
			}
		}
	}

	return stats, nil
}

// GetAllCommits retrieves all commits in the repository in reverse chronological order.
// Returns commit hashes with their timestamps.
func GetAllCommits(repoPath string) ([]CommitInfo, error) {
	// Run: git -C <repo> log --all --pretty=format:%H|%ct
	cmd := exec.Command("git", "-C", repoPath, "log", "--all", "--pretty=format:%H|%ct")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git log failed: %w", err)
	}

	var commits []CommitInfo
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "|")
		if len(parts) != 2 {
			continue
		}

		timestamp, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			continue
		}

		commits = append(commits, CommitInfo{
			Hash: parts[0],
			Date: time.Unix(timestamp, 0),
		})
	}

	return commits, nil
}

// CommitInfo represents basic information about a commit.
type CommitInfo struct {
	Hash string
	Date time.Time
}
