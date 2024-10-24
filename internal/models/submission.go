package models

import (
	"github.com/gurkengewuerz/GitCodeJudge/internal/gitea"
	"time"
)

type Status string

const (
	StatusPassed Status = "passed"
	StatusFailed Status = "failed"
	StatusError  Status = "error"
)

type ExecutionResult struct {
	Output        string
	Error         string
	ExitCode      int64
	ExecutionTime time.Duration
}

type TestCaseResult struct {
	TestNumber    int
	Solution      Solution
	Status        Status
	Error         string
	ExecutionTime time.Duration
}

type Solution struct {
	Workshop string
	Task     string
}

type Submission struct {
	RepoName   string
	CommitID   string
	BranchName string
	CloneURL   string
	Solutions  []Solution
	GitClient  *gitea.Client
}

type TestResult struct {
	Status    Status
	TestCases []TestCaseResult
}
