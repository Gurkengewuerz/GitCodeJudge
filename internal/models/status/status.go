package status

type Status string

const (
	StatusNone   Status = "none"
	StatusPassed Status = "passed"
	StatusFailed Status = "failed"
	StatusError  Status = "error"
)
