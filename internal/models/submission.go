package models

import (
	"github.com/gurkengewuerz/GitCodeJudge/internal/gitea"
	"github.com/gurkengewuerz/GitCodeJudge/internal/models/status"
	"time"
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
	Status        status.Status
	Error         string
	ExecutionTime time.Duration
	IsHidden      bool
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
	GitClient  *gitea.GiteaClient
}

type TestResult struct {
	Status    status.Status
	TestCases []TestCaseResult
	Markdown  string
}
