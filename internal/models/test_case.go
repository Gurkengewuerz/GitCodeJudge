package models

import "time"

type TestCase struct {
	Input         string
	Expected      string
	IsHidden      bool
	RepositoryDir string
	Solution      *Solution
}

type Case struct {
	Input    string `yaml:"input"`
	Expected string `yaml:"expected"`
}

type TestCaseConfig struct {
	Name        string     `yaml:"name"`
	Description string     `yaml:"description"`
	Cases       []Case     `yaml:"cases"`
	HiddenCases []Case     `yaml:"hidden_cases"`
	Disabled    bool       `default:"false" yaml:"disabled"`
	StartDate   *time.Time `yaml:"start_date"`
	EndDate     *time.Time `yaml:"end_date"`
}
