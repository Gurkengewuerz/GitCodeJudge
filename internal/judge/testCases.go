package judge

import (
	"fmt"
	"github.com/gurkengewuerz/GitCodeJudge/internal/models"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// LoadTestCases loads all test cases from the specified directory
func LoadTestCases(taskDir string) ([]models.TestCase, error) {
	// Check if directory exists
	if _, err := os.Stat(taskDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("task directory does not exist: %s", taskDir)
	}

	// Look for config.yaml first
	configPath := filepath.Join(taskDir, "config.yaml")
	if _, err := os.Stat(configPath); err == nil {
		return loadTestCasesFromConfig(configPath)
	}

	return make([]models.TestCase, 0), nil
}

// loadTestCasesFromConfig loads test cases from a YAML configuration file
func loadTestCasesFromConfig(configPath string) ([]models.TestCase, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config models.TestCaseConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	if config.Disabled {
		return nil, nil
	}

	now := time.Now()
	// If start date exists and check time is before start
	if config.StartDate != nil && now.Before(*config.StartDate) {
		return nil, nil
	}

	// If end date exists and check time is after end
	if config.EndDate != nil && now.After(*config.EndDate) {
		return nil, nil
	}

	testCases := make([]models.TestCase, len(config.Cases))
	listCases := [][]models.Case{config.Cases, config.HiddenCases}
	for j, cases := range listCases {
		for i, c := range cases {
			testCases[i] = models.TestCase{
				Input:    c.Input,
				Expected: FormatExpectedString(c.Expected),
				IsHidden: j == 1,
			}
		}
	}

	return testCases, nil
}

func FormatExpectedString(expected string) string {
	expectedLines := strings.Split(expected, "\n")

	if len(expectedLines) > 1 {
		// if the first is just a dot its used yaml
		if Trim(expectedLines[0]) == "." {
			expectedLines = expectedLines[1:]
		}
	}

	return strings.Join(expectedLines, "\n")
}
