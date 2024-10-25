package models

import (
	"fmt"
	"github.com/gurkengewuerz/GitCodeJudge/internal/models/status"
	"sort"
	"strings"
	"time"
)

func FormatTestResult(result *TestResult) string {
	var b strings.Builder

	// Write header with overall status
	switch result.Status {
	case status.StatusPassed:
		b.WriteString("## ✅ All Tests Passed\n\n")
	case status.StatusFailed:
		b.WriteString("## ❌ Some Tests Failed\n\n")
	case status.StatusError:
		b.WriteString("## ⚠️ Execution Error\n\n")
	}

	// Write detailed results for each test case
	b.WriteString("### Test Results\n\n")
	b.WriteString("| Test # | Task | Status | Time | Details |\n")
	b.WriteString("|--------|------|--------|------|----------|\n")

	for _, tc := range result.TestCases {
		resultStatus := "✅"
		if tc.Status == status.StatusFailed {
			resultStatus = "❌"
		} else if tc.Status == status.StatusError {
			resultStatus = "⚠️"
		}

		details := ""
		if tc.Error != "" {
			details = fmt.Sprintf("`%s`", tc.Error)
		}

		if tc.IsHidden {
			details = "_redacted output for hidden test_"
		}

		b.WriteString(fmt.Sprintf("| %d | %s/%s | %s | %.2fs | %s |\n",
			tc.TestNumber,
			tc.Solution.Workshop,
			tc.Solution.Task,
			resultStatus,
			tc.ExecutionTime.Seconds(),
			details))
	}

	return b.String()
}

func FormatWorkshopStats(baseURL string, workshop string, task string, stats *WorkshopStats) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("# Statistics for [%s/%s](/pdf?task=%s/%s)\n\n", workshop, task, workshop, task))

	// Overview section
	b.WriteString("## Overview\n\n")
	b.WriteString(fmt.Sprintf("- Total Completions: **%d**\n", stats.TotalUsers))
	if stats.LatestSubmit.Year() > 1 {
		b.WriteString(fmt.Sprintf("- Latest Completion: **%s**\n", stats.LatestSubmit.Format(time.RFC850)))
	}
	b.WriteString("\n")

	// Sort submissions by timestamp (most recent first)
	sort.Slice(stats.Submissions, func(i, j int) bool {
		return stats.Submissions[i].Timestamp.After(stats.Submissions[j].Timestamp)
	})

	b.WriteString("## Submissions\n\n")
	b.WriteString("| User | Completion Date | Repository | Commit |\n")
	b.WriteString("|------|-----------------|------------|--------|\n")

	for _, submission := range stats.Submissions {
		parts := strings.Split(submission.RepoName, "/")
		username := parts[1]
		b.WriteString(fmt.Sprintf("| [%s](/user/%s) | %s | [%s](%s) | [`%s`](%s/results/%s) |\n",
			username,
			username,
			submission.Timestamp.Format(time.RFC850),
			submission.RepoName,
			submission.CloneURL,
			submission.CommitID[:8],
			baseURL,
			submission.CommitID))
	}

	return b.String()
}

func FormatUserStats(baseURL string, progress *ScoreboardUserProgress) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("# Progress for %s\n\n", progress.User))

	// Task count summary
	b.WriteString(fmt.Sprintf("## Overview\n\n"))
	b.WriteString(fmt.Sprintf("Total Completed Tasks: **%d**\n\n", len(progress.Submissions)))

	// Sort submissions by timestamp (most recent first)
	sort.Slice(progress.Submissions, func(i, j int) bool {
		return progress.Submissions[i].Submission.Timestamp.After(progress.Submissions[j].Submission.Timestamp)
	})

	b.WriteString("## Completed Tasks\n\n")
	b.WriteString("| Workshop | Task | Completion Date | Repository | Commit |\n")
	b.WriteString("|----------|------|-----------------|------------|--------|\n")

	for _, submission := range progress.Submissions {
		b.WriteString(fmt.Sprintf("| %s | [%s](/workshop/%s/%s) | %s | [%s](%s) | [`%s`](%s/results/%s) |\n",
			submission.Workshop,
			submission.Task,
			submission.Workshop,
			submission.Task,
			submission.Submission.Timestamp.Format(time.RFC850),
			submission.Submission.RepoName,
			submission.Submission.CloneURL,
			submission.Submission.CommitID[:8],
			baseURL,
			submission.Submission.CommitID))
	}
	return b.String()
}

func FormatLeaderboard(leaderboard []Leaderboard) string {
	var b strings.Builder
	b.WriteString("# 🏆 Leaderboard\n\n")

	// Add total participants info if available
	if len(leaderboard) > 0 {
		b.WriteString(fmt.Sprintf("Showing top %d participants\n\n", len(leaderboard)))
	}

	b.WriteString("| Rank | User | Completed Tasks | Latest Submission | Latest Repository |\n")
	b.WriteString("|------|------|-----------------|-------------------|------------------|\n")

	for i, entry := range leaderboard {
		b.WriteString(fmt.Sprintf("| %d | [%s](/user/%s) | %d | %s | %s |\n",
			i+1,
			entry.Username,
			entry.Username,
			entry.CompletedTasks,
			entry.LastSubmission.Format(time.RFC850),
			entry.LatestRepoName))
	}

	return b.String()
}
