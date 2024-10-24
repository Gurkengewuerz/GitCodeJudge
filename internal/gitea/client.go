package gitea

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

type IssueRequest struct {
	Title  string `json:"title"`
	Body   string `json:"body"`
	Closed bool   `json:"closed"`
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

func (c *Client) CreateCommitComment(owner, repo, commit string, status string, comment string) error {
	url := fmt.Sprintf("%s/api/v1/repos/%s/%s/issues", c.baseURL, owner, repo)

	// Create issue request with the status prefix in title
	reqBody := IssueRequest{
		Title:  fmt.Sprintf("[%s] Test Results for commit %.7s", strings.ToUpper(string(status)), commit),
		Body:   fmt.Sprintf("Test results for commit: %s\n\n%s", commit, comment),
		Closed: true, // Create the issue as closed
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal comment: %v", err)
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
