package models

import (
	"fmt"
	"github.com/gurkengewuerz/GitCodeJudge/internal/models/status"
	"strings"
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
