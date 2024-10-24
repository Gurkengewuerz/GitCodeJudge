package judge

import (
	"fmt"
	"github.com/gurkengewuerz/GitCodeJudge/internal/models"
	"gopkg.in/yaml.v3"
	"os"
	"path"
	"path/filepath"
	"strings"
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

	// Fallback to loading from separate files if no config.yaml exists
	return loadTestCasesFromFiles(taskDir)
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

// loadTestCasesFromFiles loads test cases from separate input/output files
func loadTestCasesFromFiles(taskDir string) ([]models.TestCase, error) {
	// Read directory entries
	entries, err := os.ReadDir(taskDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read task directory: %v", err)
	}

	// Map to store input/output file pairs
	inputFiles := make(map[string]string)
	outputFiles := make(map[string]string)

	if _, err := os.Stat(path.Join(taskDir, ".disabled")); err == nil {
		return nil, nil
	}

	// Categorize files
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if strings.HasPrefix(name, "input") && strings.HasSuffix(name, ".txt") {
			// Extract test number (e.g., "input1.txt" -> "1")
			num := strings.TrimPrefix(strings.TrimSuffix(name, ".txt"), "input")
			inputFiles[num] = filepath.Join(taskDir, name)
		} else if strings.HasPrefix(name, "output") && strings.HasSuffix(name, ".txt") {
			num := strings.TrimPrefix(strings.TrimSuffix(name, ".txt"), "output")
			outputFiles[num] = filepath.Join(taskDir, name)
		}
	}

	// Match input/output pairs and create test cases
	var testCases []models.TestCase
	for num, inputPath := range inputFiles {
		outputPath, exists := outputFiles[num]
		if !exists {
			return nil, fmt.Errorf("missing output file for input%s.txt", num)
		}

		// Read input file
		input, err := os.ReadFile(inputPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read input file %s: %v", inputPath, err)
		}

		// Read output file
		expected, err := os.ReadFile(outputPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read output file %s: %v", outputPath, err)
		}

		testCases = append(testCases, models.TestCase{
			Input:    string(input),
			Expected: string(expected),
		})
	}

	if len(testCases) == 0 {
		return nil, fmt.Errorf("no test cases found in directory: %s", taskDir)
	}

	return testCases, nil
}
