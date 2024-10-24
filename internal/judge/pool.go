package judge

import (
	"github.com/gurkengewuerz/GitCodeJudge/internal/models"
	"log"
	"sync"
)

type Pool struct {
	executor    *Executor
	maxWorkers  int
	workers     chan struct{}
	submissions chan models.Submission
	wg          sync.WaitGroup
}

func NewPool(executor *Executor, maxWorkers int) *Pool {
	p := &Pool{
		executor:    executor,
		maxWorkers:  maxWorkers,
		workers:     make(chan struct{}, maxWorkers),
		submissions: make(chan models.Submission, 1000), // Buffer for pending submissions
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
		result, err := p.executor.Execute(submission)
		if err != nil {
			log.Printf("Failed to execute submission by %s @ %s: %v", submission.RepoName, submission.CommitID, err)
			continue
		}

		if len(result.TestCases) > 0 {
			if err := PostResultToGitea(submission, result); err != nil {
				log.Printf("Failed to post result: %v", err)
			} else {
				log.Printf("Posting Results to %s status %s", submission.RepoName, result.Status)
			}
		} else {
			log.Printf("No solutions found in submission by %s @ %s: %v", submission.RepoName, submission.CommitID, err)
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
