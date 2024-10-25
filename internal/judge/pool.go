package judge

import (
	"bytes"
	"fmt"
	"github.com/dgraph-io/badger/v4"
	appConfig "github.com/gurkengewuerz/GitCodeJudge/config"
	"github.com/gurkengewuerz/GitCodeJudge/db"
	"github.com/gurkengewuerz/GitCodeJudge/internal/judge/scoreboard"
	"github.com/gurkengewuerz/GitCodeJudge/internal/models"
	"github.com/gurkengewuerz/GitCodeJudge/internal/models/status"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"log"
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
	mdParser          goldmark.Markdown
	scoreboardManager *scoreboard.ScoreboardManager
}

func NewPool(executor *Executor, scoreboardManager *scoreboard.ScoreboardManager, maxWorkers int) *Pool {
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
		),
	)

	p := &Pool{
		executor:          executor,
		maxWorkers:        maxWorkers,
		workers:           make(chan struct{}, maxWorkers),
		submissions:       make(chan models.Submission, 1000), // Buffer for pending submissions
		mdParser:          md,
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
		// Extract owner and repo from full repository name
		parts := strings.Split(submission.RepoName, "/")
		if len(parts) != 2 {
			log.Printf("invalid repository name format: %s", submission.RepoName)
			continue
		}

		owner, repo := parts[0], parts[1]
		targetURL := fmt.Sprintf("%s/results/%s", submission.BaseURL, submission.CommitID)

		if err := submission.GitClient.PostStarting(owner, repo, submission.CommitID, targetURL, status.StatusNone, "Judge started"); err != nil {
			log.Printf("Failed to post starting: %v", err)
		} else {
			log.Printf("Posting starting to %s", submission.RepoName)
		}

		result, err := p.executor.Execute(submission)
		if err != nil {
			log.Printf("Failed to execute submission by %s @ %s: %v", submission.RepoName, submission.CommitID, err)

			if err := submission.GitClient.PostStarting(owner, repo, submission.CommitID, targetURL, status.StatusError, "Internal Server error"); err != nil {
				log.Printf("Failed to post internal server error: %v", err)
			} else {
				log.Printf("Posting internal server error %s", submission.RepoName)
			}

			continue
		}

		var markdownBuf bytes.Buffer
		markdown := models.FormatTestResult(result)
		if err := p.mdParser.Convert([]byte(markdown), &markdownBuf); err != nil {
			log.Printf("Failed to generate markdown results by %s @ %s: %v", submission.RepoName, submission.CommitID, err)
		}
		result.Markdown = markdownBuf.String()

		err = db.DB.Update(func(txn *badger.Txn) error {
			e := badger.NewEntry([]byte(submission.CommitID), markdownBuf.Bytes())
			if appConfig.CFG.DatabaseTTL != 0 {
				e = e.WithTTL(time.Hour * time.Duration(appConfig.CFG.DatabaseTTL))
			}
			err := txn.SetEntry(e)
			return err
		})
		if err != nil {
			log.Printf("Failed to create database entry %s @ %s: %v", submission.RepoName, submission.CommitID, err)
		} else {
			log.Printf("Created Results in database %s @ %s", submission.RepoName, submission.CommitID)
		}

		if len(result.TestCases) == 0 {
			log.Printf("No solutions found in submission by %s @ %s: %v", submission.RepoName, submission.CommitID, err)
			result.Status = status.StatusNone
		} else {
			if err := p.scoreboardManager.ProcessTestResults(submission, result.TestCases); err != nil {
				log.Printf("Failed to process test results for scoreboard: %v", err)
			}
		}

		if err := submission.GitClient.PostResult(owner, repo, submission.CommitID, targetURL, result.Status); err != nil {
			log.Printf("Failed to post result: %v", err)
		} else {
			log.Printf("Posting Results to %s status %s", submission.RepoName, result.Status)
		}
	}
}

func (p *Pool) Submit(submission models.Submission) {
	p.submissions <- submission
}

func (p *Pool) Stop() {
	close(p.submissions)
	p.wg.Wait()
}
