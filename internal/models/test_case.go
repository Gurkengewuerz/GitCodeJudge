package models

type TestCase struct {
	Input         string
	Expected      string
	RepositoryDir string
	Solution      *Solution
}
