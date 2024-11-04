// testCases.go

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

// WorkshopTask represents a workshop task with its configuration and path information
type WorkshopTask struct {
	Workshop   string                `json:"workshop"`
	Task       string                `json:"task"`
	Config     models.TestCaseConfig `json:"config"`
	ConfigPath string                `json:"config_path"`
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

// FindAllTasks finds and loads all available workshop tasks
func FindAllTasks(testPath string) ([]WorkshopTask, error) {
	var tasks []WorkshopTask

	err := filepath.WalkDir(testPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.Name() == "config.yaml" {
			// Read and parse config file
			yamlData, err := os.ReadFile(path)
			if err != nil {
				return nil // Skip this file if we can't read it
			}

			var config models.TestCaseConfig
			if err := yaml.Unmarshal(yamlData, &config); err != nil {
				return nil // Skip this file if we can't parse it
			}

			// Get relative path components
			relPath, err := filepath.Rel(testPath, filepath.Dir(path))
			if err != nil {
				return nil // Skip if we can't get relative path
			}

			pathParts := strings.Split(relPath, string(os.PathSeparator))
			if len(pathParts) != 2 {
				return nil // Skip if path structure is not workshop/task
			}

			tasks = append(tasks, WorkshopTask{
				Workshop:   pathParts[0],
				Task:       pathParts[1],
				Config:     config,
				ConfigPath: path,
			})
		}

		return nil
	})

	return tasks, err
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

	listCases := [][]models.Case{config.Cases, config.HiddenCases}
	testCases := make([]models.TestCase, 0)

	for j, cases := range listCases {
		for _, c := range cases {
			testCases = append(testCases, models.TestCase{
				Input:    c.Input,
				Expected: FormatExpectedString(c.Expected),
				IsHidden: j == 1,
			})
		}
	}

	return testCases, nil
}

// LoadWorkshopTask loads the configuration for a specific workshop task
func LoadWorkshopTask(testPath, workshopID, taskID string) (*WorkshopTask, error) {
	// Clean and validate path components
	workshopID = filepath.Clean(workshopID)
	taskID = filepath.Clean(taskID)

	if strings.Contains(workshopID, "..") || strings.Contains(taskID, "..") {
		return nil, fmt.Errorf("invalid path components")
	}

	configPath := filepath.Join(testPath, workshopID, taskID, "config.yaml")

	// Check if config file exists
	if _, err := os.Stat(configPath); err != nil {
		return nil, fmt.Errorf("config file not found: %v", err)
	}

	// Read and parse config file
	yamlData, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config models.TestCaseConfig
	if err := yaml.Unmarshal(yamlData, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	// Create and return WorkshopTask
	task := &WorkshopTask{
		Workshop:   workshopID,
		Task:       taskID,
		Config:     config,
		ConfigPath: configPath,
	}

	return task, nil
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
