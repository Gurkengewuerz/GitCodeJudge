package gitea

import (
	"code.gitea.io/sdk/gitea"
	"fmt"
	"github.com/gurkengewuerz/GitCodeJudge/internal/models/status"
	log "github.com/sirupsen/logrus"
)

type GiteaClient struct {
	baseURL string
	token   string
	client  *gitea.Client
}

func NewGiteaClient(baseURL, token string) *GiteaClient {
	client, err := gitea.NewClient(baseURL, gitea.SetToken(token))
	if err != nil {
		return nil
	}

	return &GiteaClient{
		baseURL: baseURL,
		token:   token,
		client:  client,
	}
}

func (c *GiteaClient) PostStarting(owner string, repo string, commit string, targetURL string, resStatus status.Status, comment string) error {
	state := gitea.StatusPending
	if resStatus == status.StatusNone {
		state = gitea.StatusPending
	} else if resStatus == status.StatusPassed {
		state = gitea.StatusSuccess
	} else if resStatus == status.StatusFailed {
		state = gitea.StatusFailure
	} else if resStatus == status.StatusError {
		state = gitea.StatusError
	}

	return c.createCommitStatus(owner, repo, commit, targetURL, state, comment)
}

func (c *GiteaClient) PostResult(owner string, repo string, commit string, targetURL string, resStatus status.Status) error {
	state := gitea.StatusSuccess
	comment := "Judge successful"
	if resStatus == status.StatusNone {
		state = gitea.StatusSuccess
		comment = "No Testcases found"
	} else if resStatus == status.StatusFailed {
		state = gitea.StatusFailure
		comment = "Judge failed"
	} else if resStatus == status.StatusError {
		state = gitea.StatusError
		comment = "Judge error"
	}

	return c.createCommitStatus(owner, repo, commit, targetURL, state, comment)
}

func (c *GiteaClient) createCommitStatus(owner, repo, sha string, targetURL string, status gitea.StatusState, description string) error {
	option := gitea.CreateStatusOption{
		State:       status,
		TargetURL:   targetURL,
		Description: description,
		Context:     "continuous-integration/judge", // You can customize this context
	}

	log.WithFields(log.Fields{
		"Body": option,
	}).Debug("Sending via Gitea client")

	_, resp, err := c.client.CreateStatus(owner, repo, sha, option)

	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	defer resp.Body.Close()

	return nil
}
