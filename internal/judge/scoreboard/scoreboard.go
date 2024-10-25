package scoreboard

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgraph-io/badger/v4"
	appConfig "github.com/gurkengewuerz/GitCodeJudge/config"
	"github.com/gurkengewuerz/GitCodeJudge/internal/models"
	"github.com/gurkengewuerz/GitCodeJudge/internal/models/status"
	log "github.com/sirupsen/logrus"
	"sort"
	"strings"
	"time"
)

type ScoreboardManager struct {
	db *badger.DB
}

func NewScoreboardManager(db *badger.DB) *ScoreboardManager {
	log.Info("New Scoreboard Manager created")

	return &ScoreboardManager{db: db}
}

func (sm *ScoreboardManager) ProcessTestResults(submission models.Submission, testCases []models.TestCaseResult) error {
	parts := strings.Split(submission.RepoName, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid repository name format: %s", submission.RepoName)
	}
	username := parts[1]

	// Group test cases by workshop/task
	taskResults := make(map[models.ScoreboardWorkshopTask]bool)
	for _, tc := range testCases {
		wt := models.ScoreboardWorkshopTask{
			Workshop: tc.Solution.Workshop,
			Task:     tc.Solution.Task,
		}

		if _, exists := taskResults[wt]; !exists {
			taskResults[wt] = true
		}

		if tc.Status != status.StatusPassed {
			taskResults[wt] = false
		}
	}

	return sm.db.Update(func(txn *badger.Txn) error {
		// Process each passed workshop/task
		for wt, passed := range taskResults {
			if !passed {
				continue
			}

			userSubmission := models.ScoreboardUserSubmission{
				RepoName:  submission.RepoName,
				CommitID:  submission.CommitID,
				CloneURL:  submission.CloneURL,
				Timestamp: time.Now(),
			}

			if err := sm.updateUserProgress(txn, username, wt, userSubmission); err != nil {
				return err
			}

			if err := sm.updateWorkshopStats(txn, wt, userSubmission); err != nil {
				return err
			}
		}
		return nil
	})
}

func (sm *ScoreboardManager) updateUserProgress(txn *badger.Txn, username string, wt models.ScoreboardWorkshopTask, submission models.ScoreboardUserSubmission) error {
	userKey := []byte(fmt.Sprintf("user:%s", username))
	var progress models.ScoreboardUserProgress

	item, err := txn.Get(userKey)
	if err != nil && !errors.Is(err, badger.ErrKeyNotFound) {
		return err
	}

	if err == nil {
		err = item.Value(func(val []byte) error {
			return json.Unmarshal(val, &progress)
		})
		if err != nil {
			return err
		}
	} else {
		progress = models.ScoreboardUserProgress{
			User: username,
			Submissions: []struct {
				Workshop   string                          `json:"workshop"`
				Task       string                          `json:"task"`
				Submission models.ScoreboardUserSubmission `json:"submission"`
			}{},
		}
	}

	// Check if this workshop/task already exists
	found := false
	for i, s := range progress.Submissions {
		if s.Workshop == wt.Workshop && s.Task == wt.Task {
			progress.Submissions[i].Submission = submission
			found = true
			break
		}
	}

	if !found {
		progress.Submissions = append(progress.Submissions, struct {
			Workshop   string                          `json:"workshop"`
			Task       string                          `json:"task"`
			Submission models.ScoreboardUserSubmission `json:"submission"`
		}{
			Workshop:   wt.Workshop,
			Task:       wt.Task,
			Submission: submission,
		})
	}

	data, err := json.Marshal(progress)
	if err != nil {
		return err
	}

	e := badger.NewEntry(userKey, data)
	if appConfig.CFG.DatabaseTTL != 0 {
		e = e.WithTTL(time.Hour * time.Duration(appConfig.CFG.DatabaseTTL))
	}
	return txn.SetEntry(e)
}

func (sm *ScoreboardManager) updateWorkshopStats(txn *badger.Txn, wt models.ScoreboardWorkshopTask, submission models.ScoreboardUserSubmission) error {
	workshopKey := []byte(fmt.Sprintf("workshop:%s:%s", wt.Workshop, wt.Task))
	var stats models.WorkshopStats

	item, err := txn.Get(workshopKey)
	if err != nil && !errors.Is(err, badger.ErrKeyNotFound) {
		return err
	}

	if err == nil {
		err = item.Value(func(val []byte) error {
			return json.Unmarshal(val, &stats)
		})
		if err != nil {
			return err
		}
	}

	stats.TotalUsers++
	stats.CompletedAt = append(stats.CompletedAt, time.Now())
	stats.LatestSubmit = time.Now()
	stats.Submissions = append(stats.Submissions, submission)

	data, err := json.Marshal(stats)
	if err != nil {
		return err
	}

	e := badger.NewEntry(workshopKey, data)
	if appConfig.CFG.DatabaseTTL != 0 {
		e = e.WithTTL(time.Hour * time.Duration(appConfig.CFG.DatabaseTTL))
	}
	return txn.SetEntry(e)
}

func (sm *ScoreboardManager) GetUserProgress(username string) (*models.ScoreboardUserProgress, error) {
	var progress models.ScoreboardUserProgress

	err := sm.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(fmt.Sprintf("user:%s", username)))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &progress)
		})
	})

	if errors.Is(err, badger.ErrKeyNotFound) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &progress, nil
}

func (sm *ScoreboardManager) GetWorkshopStats(workshop, task string) (*models.WorkshopStats, error) {
	var stats models.WorkshopStats

	err := sm.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(fmt.Sprintf("workshop:%s:%s", workshop, task)))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &stats)
		})
	})

	if errors.Is(err, badger.ErrKeyNotFound) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &stats, nil
}

func (sm *ScoreboardManager) GetLeaderboard(limit int) ([]models.Leaderboard, error) {
	type userScore struct {
		username       string
		completedTasks int
		lastSubmission time.Time
		latestRepoName string
	}

	var scores []userScore

	err := sm.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte("user:")

		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			var progress models.ScoreboardUserProgress

			err := item.Value(func(val []byte) error {
				return json.Unmarshal(val, &progress)
			})
			if err != nil {
				return err
			}

			var lastSubmission time.Time
			var latestRepoName string

			for _, sub := range progress.Submissions {
				if sub.Submission.Timestamp.After(lastSubmission) {
					lastSubmission = sub.Submission.Timestamp
					latestRepoName = sub.Submission.RepoName
				}
			}

			scores = append(scores, userScore{
				username:       progress.User,
				completedTasks: len(progress.Submissions),
				lastSubmission: lastSubmission,
				latestRepoName: latestRepoName,
			})
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort scores
	sort.Slice(scores, func(i, j int) bool {
		if scores[i].completedTasks == scores[j].completedTasks {
			return scores[i].lastSubmission.After(scores[j].lastSubmission)
		}
		return scores[i].completedTasks > scores[j].completedTasks
	})

	if limit > len(scores) {
		limit = len(scores)
	}

	result := make([]models.Leaderboard, limit)

	for i := 0; i < limit; i++ {
		result[i] = models.Leaderboard{
			Username:       scores[i].username,
			CompletedTasks: scores[i].completedTasks,
			LastSubmission: scores[i].lastSubmission,
			LatestRepoName: scores[i].latestRepoName,
		}
	}

	return result, nil
}
