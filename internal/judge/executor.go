package judge

import (
	"context"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/gurkengewuerz/GitCodeJudge/config"
	"github.com/gurkengewuerz/GitCodeJudge/internal/models"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type Executor struct {
	docker      *DockerExecutor
	testCaseDir string
}

func NewExecutor(docker *DockerExecutor, testCaseDir string) *Executor {
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
			ExtraCutset += string(i)
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

	log.Printf("Worker executes %s", submission.RepoName)

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

	log.Printf("Worker checked out %s to %s", submission.CloneURL, repoTmpDir)

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

	log.Printf("Found %d test cases", len(testCases))

	result := &models.TestResult{
		TestCases: make([]models.TestCaseResult, len(testCases)),
	}

	// Run each test case
	for i, tc := range testCases {
		log.Printf("Executing %s/%s for %s now", tc.Solution.Workshop, tc.Solution.Task, submission.RepoName)
		execResult, err := e.docker.RunCode(context.Background(), tc)
		if err != nil {
			return nil, fmt.Errorf("failed to execute test case %d: %v", i+1, err)
		}

		caseResult := models.TestCaseResult{
			TestNumber:    i + 1,
			ExecutionTime: execResult.ExecutionTime,
			Error:         execResult.Error,
			Solution:      *tc.Solution,
		}

		log.Printf("%s/%s for %s status: %s output: %s", tc.Solution.Workshop, tc.Solution.Task, submission.RepoName, caseResult.Status, execResult.Output)
		if execResult.Error != "" {
			caseResult.Status = models.StatusError
			log.Printf("%s/%s for %s error: %s", tc.Solution.Workshop, tc.Solution.Task, submission.RepoName, caseResult.Error)
		} else if execResult.ExitCode != 0 {
			caseResult.Status = models.StatusError
			caseResult.Error = fmt.Sprintf("Program exited with code %d", execResult.ExitCode)
			caseResult.Status = models.StatusError
			log.Printf("%s/%s for %s error: %s", tc.Solution.Workshop, tc.Solution.Task, submission.RepoName, caseResult.Error)
		} else {
			// Compare output
			expectedLines := strings.Split(Trim(tc.Expected), "\n")
			actualLines := strings.Split(Trim(execResult.Output), "\n")

			if len(expectedLines) != len(actualLines) {
				caseResult.Status = models.StatusFailed
				caseResult.Error = fmt.Sprintf("Expected %d lines, got %d", len(expectedLines), len(actualLines))
			} else {
				caseResult.Status = models.StatusPassed
				for j := range expectedLines {
					expected := Trim(expectedLines[j])
					actual := Trim(actualLines[j])

					log.Printf("Testing %s/%s/%s  \"%v\" - \"%v\"", submission.RepoName, tc.Solution.Workshop, tc.Solution.Task, []rune(expected), []rune(actual))

					if expected != actual {
						caseResult.Status = models.StatusFailed
						caseResult.Error = fmt.Sprintf("Line %d mismatch: Expected: %s Got: %s", j+1, expected, actual)
						break
					}
				}
			}
		}

		result.TestCases[i] = caseResult
	}

	// Calculate overall result
	result.Status = models.StatusPassed
	for _, tc := range result.TestCases {
		if tc.Status != models.StatusPassed {
			result.Status = models.StatusFailed
			break
		}
	}

	return result, nil
}
