package judge

import (
	"bytes"
	"context"
	"fmt"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/gurkengewuerz/GitCodeJudge/internal/config"
	"github.com/gurkengewuerz/GitCodeJudge/internal/models"
	log "github.com/sirupsen/logrus"
	"io"
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
	log.Info("New Docker executer created")

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
	tmpDir, err := getTempDir("judge-*")
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

	// Check if image exists locally
	images, err := e.cli.ImageList(ctx, image.ListOptions{})
	if err != nil {
		panic(err)
	}

	imageExists := false
	for _, img := range images {
		for _, tag := range img.RepoTags {
			if tag == config.CFG.DockerImage {
				imageExists = true
				break
			}
		}
	}

	imageFields := log.Fields{
		"Image": config.CFG.DockerImage,
	}
	// Pull image if it doesn't exist
	if !imageExists {
		log.WithFields(imageFields).Warn("Image not found locally, pulling...")
		reader, err := e.cli.ImagePull(ctx, config.CFG.DockerImage, image.PullOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to pull image: %v", err)
		}
		// Wait for the image pull to complete
		_, err = io.Copy(io.Discard, reader)
		if err != nil {
			panic(err)
		}
		reader.Close()
		log.WithFields(imageFields).Info("Successfully pulled")
	} else {
		log.WithFields(imageFields).Debug("Image found locally")
	}

	// Create container
	resp, err := e.cli.ContainerCreate(ctx,
		&container.Config{
			Image: config.CFG.DockerImage,
			Env: []string{
				fmt.Sprintf("JUDGE_WORKSHOP=%s", Trim(testCase.Solution.Workshop)),
				fmt.Sprintf("JUDGE_TASK=%s", Trim(testCase.Solution.Task)),
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
	containerFields := log.Fields{
		"ContainerID": resp.ID,
	}
	defer func(cli *client.Client, ctx context.Context, containerID string, options container.RemoveOptions) {
		err := cli.ContainerRemove(ctx, containerID, options)
		if err != nil {
			log.WithFields(containerFields).WithError(err).Error("Failed to remove container")
		}
	}(e.cli, ctx, resp.ID, container.RemoveOptions{
		Force: true,
	})

	// Start container with timeout
	start := time.Now()
	if err := e.cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return nil, fmt.Errorf("failed to start container: %v", err)
	}

	log.WithFields(containerFields).WithError(err).Debug("Starting container")

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
			log.WithFields(containerFields).Error(str)
			return &result, nil
		}
	case <-ctx.Done():
		str := "execution timeout"
		result.Error = str
		log.WithFields(containerFields).Warn(str)
		return &result, nil
	case status := <-statusCh:
		if status.Error != nil {
			str := fmt.Sprintf("container error: %s", status.Error.Message)
			result.Error = str
			log.WithFields(containerFields).Error(str)
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
