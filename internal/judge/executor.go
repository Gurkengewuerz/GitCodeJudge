package judge

import (
	"context"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/gurkengewuerz/GitCodeJudge/config"
	"github.com/gurkengewuerz/GitCodeJudge/internal/models"
	"github.com/gurkengewuerz/GitCodeJudge/internal/models/status"
	log "github.com/sirupsen/logrus"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type Executor struct {
	docker      *DockerExecutor
	testCaseDir string
}

func NewExecutor(docker *DockerExecutor, testCaseDir string) *Executor {
	log.Info("New executer created")

	return &Executor{
		docker:      docker,
		testCaseDir: testCaseDir,
	}
}

var (
	ExtraCutset = ""
)

func Trim(s string) string {
	if ExtraCutset == "" {
		for i := 0; i < 32; i++ {
			ExtraCutset += string(rune(i))
		}
		ExtraCutset += "\r\n"
	}
	return strings.TrimSpace(strings.Trim(strings.TrimSpace(s), ExtraCutset))
}

func (e *Executor) Execute(submission models.Submission) (*models.TestResult, error) {
	repoTmpDir, err := os.MkdirTemp("", "jrepo-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(repoTmpDir)

	field := log.Fields{
		"Repo":   submission.RepoName,
		"Commit": submission.CommitID,
	}
	log.WithFields(field).Debug("Worker executes")

	r, err := git.PlainClone(repoTmpDir, false, &git.CloneOptions{
		URL: submission.CloneURL,
		Auth: &http.BasicAuth{
			Username: "git-judge-system", // yes, this can be anything except an empty string
			Password: config.CFG.GiteaToken,
		},
		ReferenceName:     plumbing.ReferenceName(submission.BranchName),
		Depth:             1,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to clone repo %s: %v", submission.CloneURL, err)
	}

	w, err := r.Worktree()
	if err != nil {
		return nil, fmt.Errorf("failed to get worktree %s: %v", submission.CloneURL, err)
	}

	err = w.Checkout(&git.CheckoutOptions{
		Hash: plumbing.NewHash(submission.CommitID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to checkout %s @ %s: %v", submission.CloneURL, submission.CommitID, err)
	}

	field = log.Fields{
		"Repo":     submission.RepoName,
		"Commit":   submission.CommitID,
		"CloneURL": submission.CloneURL,
		"Dir":      repoTmpDir,
	}
	log.WithFields(field).Debug("Worker checked out repository")

	testCases := make([]models.TestCase, 0)

	err = filepath.WalkDir(repoTmpDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip if it's not a directory
		if !d.IsDir() {
			return nil
		}

		// Calculate relative path
		relPath, err := filepath.Rel(repoTmpDir, path)
		if err != nil {
			return err
		}

		if strings.Count(relPath, string(os.PathSeparator)) > 2 {
			return fs.SkipDir
		}

		// Get test cases for the task
		newTestCases, err := LoadTestCases(filepath.Join(e.testCaseDir, relPath))

		if err == nil {
			for i := range newTestCases {
				parts := strings.Split(relPath, string(os.PathSeparator))
				if len(parts) != 2 {
					return nil
				}

				newTestCases[i].Solution = &models.Solution{
					Workshop: parts[0],
					Task:     parts[1],
				}

				newTestCases[i].RepositoryDir = repoTmpDir

				testCases = append(testCases, newTestCases[i])
			}
		}

		return nil
	})
	log.WithFields(field).Debugf("Found %d test cases", len(testCases))

	result := &models.TestResult{
		TestCases: make([]models.TestCaseResult, len(testCases)),
	}

	// Run each test case
	for i, tc := range testCases {
		tcField := log.Fields{
			"Workshop": tc.Solution.Workshop,
			"Task":     tc.Solution.Task,
		}
		log.WithFields(field).Debug("Executing")

		execResult, err := e.docker.RunCode(context.Background(), tc)
		if err != nil {
			return nil, fmt.Errorf("failed to execute test case %d: %v", i+1, err)
		}

		caseResult := models.TestCaseResult{
			TestNumber:    i + 1,
			ExecutionTime: execResult.ExecutionTime,
			Error:         execResult.Error,
			Solution:      *tc.Solution,
			IsHidden:      tc.IsHidden,
		}

		log.WithFields(field).WithFields(tcField).WithField("output", execResult.Output).Trace()
		if execResult.Error != "" {
			caseResult.Status = status.StatusError
			log.WithFields(field).WithFields(tcField).Error(caseResult.Error)
		} else if execResult.ExitCode != 0 {
			caseResult.Status = status.StatusError
			caseResult.Error = fmt.Sprintf("Program exited with code %d", execResult.ExitCode)
			caseResult.Status = status.StatusError
			log.WithFields(field).WithFields(tcField).Error(caseResult.Error)
		} else {
			// Compare output
			expectedLines := strings.Split(Trim(tc.Expected), "\n")
			actualLines := strings.Split(Trim(execResult.Output), "\n")

			if len(expectedLines) != len(actualLines) {
				caseResult.Status = status.StatusFailed
				caseResult.Error = fmt.Sprintf("Expected %d lines, got %d", len(expectedLines), len(actualLines))
			} else {
				caseResult.Status = status.StatusPassed
				for j := range expectedLines {
					expected := Trim(expectedLines[j])
					actual := Trim(actualLines[j])

					log.WithFields(field).WithFields(tcField).Trace(fmt.Sprintf("Testing %s/%s/%s  \"%v\" - \"%v\"", submission.RepoName, tc.Solution.Workshop, tc.Solution.Task, []rune(expected), []rune(actual)))

					if expected != actual {
						caseResult.Status = status.StatusFailed
						caseResult.Error = fmt.Sprintf("Line %d mismatch: Expected: %s Got: %s", j+1, expected, actual)
						break
					}
				}
			}
		}

		result.TestCases[i] = caseResult
	}

	// Calculate overall result
	result.Status = status.StatusPassed
	for _, tc := range result.TestCases {
		if tc.Status != status.StatusPassed {
			result.Status = status.StatusFailed
			break
		}
	}

	log.WithFields(field).Debug("Worker finished")

	return result, nil
}
