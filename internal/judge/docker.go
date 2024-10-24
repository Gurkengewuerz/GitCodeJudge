package judge

import (
	"bytes"
	"context"
	"fmt"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/gurkengewuerz/GitCodeJudge/config"
	"github.com/gurkengewuerz/GitCodeJudge/internal/models"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type DockerExecutor struct {
	cli     *client.Client
	network string
	timeout time.Duration
}

func NewDockerExecutor(network string, timeoutSeconds int) (*DockerExecutor, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %v", err)
	}

	return &DockerExecutor{
		cli:     cli,
		network: network,
		timeout: time.Duration(timeoutSeconds) * time.Second,
	}, nil
}

func (e *DockerExecutor) RunCode(ctx context.Context, testCase models.TestCase) (*models.ExecutionResult, error) {
	// Create temp directory for code and test files
	tmpDir, err := os.MkdirTemp("", "judge-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Write files
	if err := os.WriteFile(filepath.Join(tmpDir, "input.txt"), []byte(testCase.Input), 0644); err != nil {
		return nil, fmt.Errorf("failed to write input: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "expected.txt"), []byte(testCase.Expected), 0644); err != nil {
		return nil, fmt.Errorf("failed to write expected output: %v", err)
	}

	// Create container
	resp, err := e.cli.ContainerCreate(ctx,
		&container.Config{
			Image: config.CFG.DockerImage,
			Env: []string{
				fmt.Sprintf("JUDGE_WORKSHOP=%s", testCase.Solution.Workshop),
				fmt.Sprintf("JUDGE_TASK=%s", testCase.Solution.Task),
			},
			WorkingDir: "/judge",
		},
		&container.HostConfig{
			Binds: []string{
				fmt.Sprintf("%s:/judge", tmpDir),
				fmt.Sprintf("%s:/repo", testCase.RepositoryDir),
			},
			NetworkMode: container.NetworkMode(e.network),
			RestartPolicy: container.RestartPolicy{
				Name: container.RestartPolicyDisabled,
			},
			Resources: container.Resources{
				Memory:    256 * 1024 * 1024, // 256MB memory limit
				CPUPeriod: 100000,
				CPUQuota:  50000, // 0.5 CPU
			},
		}, nil, nil, "")
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %v", err)
	}
	defer func(cli *client.Client, ctx context.Context, containerID string, options container.RemoveOptions) {
		err := cli.ContainerRemove(ctx, containerID, options)
		if err != nil {
			log.Printf("Failed to remove container %v", err)
		}
	}(e.cli, ctx, resp.ID, container.RemoveOptions{
		Force: true,
	})

	// Start container with timeout
	start := time.Now()
	if err := e.cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return nil, fmt.Errorf("failed to start container: %v", err)
	}

	// Wait for container with timeout
	ctx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	statusCh, errCh := e.cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	var result models.ExecutionResult
	result.ExecutionTime = time.Since(start)

	select {
	case err := <-errCh:
		if err != nil {
			str := fmt.Sprintf("execution error: %v", err)
			result.Error = str
			log.Println(str)
			return &result, nil
		}
	case <-ctx.Done():
		str := "execution timeout"
		result.Error = str
		log.Println(str)
		return &result, nil
	case status := <-statusCh:
		if status.Error != nil {
			str := fmt.Sprintf("container error: %s", status.Error.Message)
			result.Error = str
			log.Println(str)
			return &result, nil
		}
		result.ExitCode = status.StatusCode
	}

	// Get logs
	logs, err := e.cli.ContainerLogs(context.Background(), resp.ID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get logs: %v", err)
	}
	defer logs.Close()

	// Read all output
	var output bytes.Buffer

	_, err = stdcopy.StdCopy(&output, &output, logs)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to read logs: %v", err)
	}

	result.Output = output.String()
	return &result, nil
}
