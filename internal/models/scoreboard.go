package models

import (
	"time"
)

type ScoreboardWorkshopTask struct {
	Workshop string `json:"workshop"`
	Task     string `json:"task"`
}

type ScoreboardUserSubmission struct {
	RepoName  string    `json:"repo_name"`
	CommitID  string    `json:"commit_id"`
	CloneURL  string    `json:"clone_url"`
	Timestamp time.Time `json:"timestamp"`
}

type ScoreboardUserProgress struct {
	User        string `json:"user"`
	Submissions []struct {
		Workshop   string                   `json:"workshop"`
		Task       string                   `json:"task"`
		Submission ScoreboardUserSubmission `json:"submission"`
	} `json:"submissions"`
}

type WorkshopStats struct {
	TotalUsers   int                        `json:"total_users"`
	CompletedAt  []time.Time                `json:"completed_at"`
	LatestSubmit time.Time                  `json:"latest_submit"`
	Submissions  []ScoreboardUserSubmission `json:"submissions"`
}
