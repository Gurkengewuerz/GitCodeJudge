package gitea

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gurkengewuerz/GitCodeJudge/internal/models/status"
	"net/http"
	"time"
)

type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// StatusState represents the state of a commit status
type StatusState string

const (
	StatusPending StatusState = "pending"
	StatusSuccess StatusState = "success"
	StatusError   StatusState = "error"
	StatusFailure StatusState = "failure"
)

// StatusRequest represents a status update request
type StatusRequest struct {
	State       StatusState `json:"state"`
	TargetURL   string      `json:"target_url,omitempty"`
	Description string      `json:"description"`
	Context     string      `json:"context"`
}

func NewClient(baseURL, token string) *Client {
	return &Client{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: time.Second * 10,
		},
	}
}

func (c *Client) PostStarting(owner string, repo string, commit string, targetURL string, resStatus status.Status, comment string) error {
	state := StatusPending
	if resStatus == status.StatusNone {
		state = StatusPending
	} else if resStatus == status.StatusPassed {
		state = StatusSuccess
	} else if resStatus == status.StatusFailed {
		state = StatusFailure
	} else if resStatus == status.StatusError {
		state = StatusError
	}

	return c.createCommitStatus(owner, repo, commit, targetURL, state, comment)
}

func (c *Client) PostResult(owner string, repo string, commit string, targetURL string, resStatus status.Status) error {
	state := StatusSuccess
	comment := "Judge successful"
	if resStatus == status.StatusNone {
		state = StatusSuccess
		comment = "No Testcases found"
	} else if resStatus == status.StatusFailed {
		state = StatusFailure
		comment = "Judge failed"
	} else if resStatus == status.StatusError {
		state = StatusError
		comment = "Judge error"
	}

	return c.createCommitStatus(owner, repo, commit, targetURL, state, comment)
}

func (c *Client) createCommitStatus(owner, repo, sha string, targetURL string, status StatusState, description string) error {
	url := fmt.Sprintf("%s/api/v1/repos/%s/%s/statuses/%s", c.baseURL, owner, repo, sha)

	reqBody := StatusRequest{
		State:       status,
		Description: description,
		Context:     "continuous-integration/judge", // You can customize this context
		TargetURL:   targetURL,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal status request: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("token %s", c.token))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("gitea API returned status code %d", resp.StatusCode)
	}

	return nil
}
