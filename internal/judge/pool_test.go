package judge_test

import (
	"github.com/gurkengewuerz/GitCodeJudge/internal/judge"
	"github.com/gurkengewuerz/GitCodeJudge/internal/judge/scoreboard"
	"github.com/gurkengewuerz/GitCodeJudge/internal/models"
	"testing"
)

func TestNewPool(t *testing.T) {
	executor := &judge.Executor{}
	scoreboardManager := &scoreboard.ScoreboardManager{}
	pool := judge.NewPool(executor, scoreboardManager, 5)

	if pool == nil {
		t.Fatal("Expected pool to be created")
	}
}

func TestSubmitMultiple(t *testing.T) {
	executor := &judge.Executor{}
	scoreboardManager := &scoreboard.ScoreboardManager{}
	pool := judge.NewPool(executor, scoreboardManager, 5)

	submissions := []models.Submission{
		{RepoName: "repo1", CommitID: "commit1"},
		{RepoName: "repo2", CommitID: "commit2"},
		{RepoName: "repo3", CommitID: "commit3"},
	}

	for _, submission := range submissions {
		pool.Submit(submission)
	}
}
