package judge

import (
    "fmt"
    "github.com/gurkengewuerz/GitCodeJudge/internal/models"
    "gopkg.in/yaml.v3"
    "os"
    "path/filepath"
    "time"
)

type TestCaseConfig struct {
	Cases []struct {
		Input    string `yaml:"input"`
		Expected string `yaml:"expected"`
	} `yaml:"cases"`
	Disabled  bool       `default:"false" yaml:"disabled"`
	StartDate *time.Time `yaml:"start_date"`
	EndDate   *time.Time `yaml:"end_date"`
}

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

	var config TestCaseConfig
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
	for i, c := range config.Cases {
		testCases[i] = models.TestCase{
			Input:    c.Input,
			Expected: c.Expected,
		}
	}

	return testCases, nil
}
