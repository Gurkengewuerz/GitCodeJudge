package judge

import (
	"fmt"
	"github.com/gurkengewuerz/GitCodeJudge/internal/models"
	"strings"
)

func FormatTestResult(result *models.TestResult) string {
	var b strings.Builder

	// Write header with overall status
	switch result.Status {
	case models.StatusPassed:
		b.WriteString("## ✅ All Tests Passed\n\n")
	case models.StatusFailed:
		b.WriteString("## ❌ Some Tests Failed\n\n")
	case models.StatusError:
		b.WriteString("## ⚠️ Execution Error\n\n")
	}

	// Write detailed results for each test case
	b.WriteString("### Test Results\n\n")
	b.WriteString("| Test # | Task | Status | Time | Details |\n")
	b.WriteString("|--------|------|--------|------|----------|\n")

	for _, tc := range result.TestCases {
		status := "✅"
		if tc.Status == models.StatusFailed {
			status = "❌"
		} else if tc.Status == models.StatusError {
			status = "⚠️"
		}

		details := ""
		if tc.Error != "" {
			details = fmt.Sprintf("`%s`", tc.Error)
		}

		b.WriteString(fmt.Sprintf("| %d | %s/%s | %s | %.2fs | %s |\n",
			tc.TestNumber,
			tc.Solution.Workshop,
			tc.Solution.Task,
			status,
			tc.ExecutionTime.Seconds(),
			details))
	}

	return b.String()
}

func PostResultToGitea(submission models.Submission, result *models.TestResult) error {
	comment := FormatTestResult(result)

	// Extract owner and repo from full repository name
	parts := strings.Split(submission.RepoName, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid repository name format: %s", submission.RepoName)
	}

	owner, repo := parts[0], parts[1]

	return submission.GitClient.CreateCommitStatus(owner, repo, submission.CommitID, string(result.Status), comment)
}
