package judge

import (
	"fmt"
	"github.com/dgraph-io/badger/v4"
	appConfig "github.com/gurkengewuerz/GitCodeJudge/internal/config"
	"github.com/gurkengewuerz/GitCodeJudge/internal/db"
	"github.com/gurkengewuerz/GitCodeJudge/internal/judge/scoreboard"
	"github.com/gurkengewuerz/GitCodeJudge/internal/models"
	"github.com/gurkengewuerz/GitCodeJudge/internal/models/status"
	log "github.com/sirupsen/logrus"
	"strings"
	"sync"
	"time"
)

type Pool struct {
	executor          *Executor
	maxWorkers        int
	workers           chan struct{}
	submissions       chan models.Submission
	wg                sync.WaitGroup
	scoreboardManager *scoreboard.ScoreboardManager
}

func NewPool(executor *Executor, scoreboardManager *scoreboard.ScoreboardManager, maxWorkers int) *Pool {
	log.Info("New pool created")

	p := &Pool{
		executor:          executor,
		maxWorkers:        maxWorkers,
		workers:           make(chan struct{}, maxWorkers),
		submissions:       make(chan models.Submission, 1000), // Buffer for pending submissions
		scoreboardManager: scoreboardManager,
	}

	// Start worker pool
	go p.start()
	return p
}

func (p *Pool) start() {
	for i := 0; i < p.maxWorkers; i++ {
		p.wg.Add(1)
		go p.worker()
	}
}

func (p *Pool) worker() {
	defer p.wg.Done()

	for submission := range p.submissions {
		fields := log.Fields{
			"Repo":   submission.RepoName,
			"Commit": submission.CommitID,
		}

		// Extract owner and repo from full repository name
		parts := strings.Split(submission.RepoName, "/")
		if len(parts) != 2 {
			log.WithFields(fields).Error("invalid repository name format")
			continue
		}

		owner, repo := parts[0], parts[1]
		targetURL := fmt.Sprintf("%s/results/%s", submission.BaseURL, submission.CommitID)

		if err := submission.GitClient.PostStarting(owner, repo, submission.CommitID, targetURL, status.StatusNone, "Judge started"); err != nil {
			log.WithFields(fields).WithError(err).Error("Failed to post starting")
		} else {
			log.WithFields(fields).Info("Posting starting")
		}

		result, err := p.executor.Execute(submission)
		if err != nil {
			log.WithFields(fields).WithError(err).Error("Failed to execute submission")
			if err := submission.GitClient.PostStarting(owner, repo, submission.CommitID, targetURL, status.StatusError, "Internal Server error"); err != nil {
				log.WithFields(fields).WithError(err).Error("Failed to post internal server error")
			} else {
				log.WithFields(fields).Info("Posting internal server error")
			}
			continue
		}

		result.Markdown = models.FormatTestResult(result)

		log.WithFields(fields).Debug("Inserting results of commit to datbase")
		err = db.DB.Update(func(txn *badger.Txn) error {
			e := badger.NewEntry([]byte(submission.CommitID), []byte(result.Markdown))
			if appConfig.CFG.DatabaseTTL != 0 {
				e = e.WithTTL(time.Hour * time.Duration(appConfig.CFG.DatabaseTTL))
			}
			err := txn.SetEntry(e)
			return err
		})
		if err != nil {
			log.WithFields(fields).WithError(err).Error("Failed to create database entry")
		} else {
			log.WithFields(fields).Debug("Created Results in database")
		}

		if len(result.TestCases) == 0 {
			log.WithFields(fields).WithError(err).Warn("No solutions found in submission")
			result.Status = status.StatusNone
		} else {
			if err := p.scoreboardManager.ProcessTestResults(submission, result.TestCases); err != nil {
				log.WithFields(fields).WithError(err).Error("Failed to process test results for scoreboard")
			} else {
				log.WithFields(fields).Debug("Processed scoreboard results in database")
			}
		}

		if err := submission.GitClient.PostResult(owner, repo, submission.CommitID, targetURL, result.Status); err != nil {
			log.WithFields(fields).WithError(err).Error("Failed to post result")
		} else {
			log.WithFields(fields).Info("Posting results")
		}
	}
}

func (p *Pool) Submit(submission models.Submission) {
	p.submissions <- submission
	log.Info("Submission added")
}

func (p *Pool) Stop() {
	log.Info("Stopping pool")
	close(p.submissions)
	p.wg.Wait()
}
