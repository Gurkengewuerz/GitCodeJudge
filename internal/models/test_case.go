package models

import "time"

type TestCase struct {
	Input         string
	Expected      string
	RepositoryDir string
	Solution      *Solution
}

type TestCaseConfig struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Cases       []struct {
		Input    string `yaml:"input"`
		Expected string `yaml:"expected"`
	} `yaml:"cases"`
	Disabled  bool       `default:"false" yaml:"disabled"`
	StartDate *time.Time `yaml:"start_date"`
	EndDate   *time.Time `yaml:"end_date"`
}
