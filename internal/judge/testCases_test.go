package judge_test

import (
	"github.com/gurkengewuerz/GitCodeJudge/internal/judge"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadTestCases(t *testing.T) {
	// Define the root directory containing task directories with config.yaml files
	rootDir := "../../test_cases/"

	// Walk through the root directory to find all config.yaml files
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check if the file is config.yaml
		if info.Name() == "config.yaml" {
			t.Run(path, func(t *testing.T) {
				testCases, err := judge.LoadTestCases(filepath.Dir(path))
				if err != nil {
					t.Fatalf("Failed to load test cases from %s: %v", path, err)
				}

				// Perform some basic checks on the loaded test cases
				if len(testCases) == 0 {
					t.Errorf("No test cases loaded from %s", path)
				}

				for _, testCase := range testCases {
					if testCase.Input == "" {
						t.Errorf("Test case with empty input in %s", path)
					}
					if testCase.Expected == "" {
						t.Errorf("Test case with empty expected output in %s", path)
					}
				}
			})
		}

		return nil
	})

	cwd, cerr := os.Getwd()
	if cerr != nil {
		t.Fatalf("Failed to get cwd: %v", cerr)
	}

	if err != nil {
		t.Fatalf("Failed to walk through root directory %s: %v", cwd, err)
	}
}
