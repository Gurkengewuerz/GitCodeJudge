package judge

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strings"
)

var (
	isDocker       bool
	dockerRepoPath string
	containerID    string
)

// init runs when the package is initialized
func init() {
	var err error
	isDocker, containerID, err = detectDocker()
	if err != nil {
		log.WithError(err).Warn("Failed to check if running in Docker")
		isDocker = false
		return
	}

	if isDocker {
		// Get the mount path using the container ID
		dockerRepoPath, err = GetContainerInfo("/repos")
		if err != nil {
			log.WithError(err).Fatal("Failed to find Docker volume path for /repos")
			panic("Docker volume /repos must be mounted")
		}
		log.WithFields(log.Fields{
			"isDocker":       isDocker,
			"containerID":    containerID,
			"dockerRepoPath": dockerRepoPath,
		}).Info("Docker environment detected")
	}
}

// getTempDir returns the appropriate temporary directory path based on the environment
func getTempDir(prefix string) (string, error) {
	if isDocker {
		// When in Docker, create temp dir under the host path
		containerPath := filepath.Join("/repos", strings.ReplaceAll(prefix, "*", fmt.Sprintf("%d", rand.IntN(100000))))
		err := os.Mkdir(containerPath, 0755)
		if err != nil {
			return "", err
		}
		return containerPath, nil
	}
	// Default behavior when not in Docker
	return os.MkdirTemp("", prefix)
}

func getHostPath(localPath string) string {
	if !isDocker {
		return localPath
	}
	dir := filepath.Base(localPath)
	hostDir := fmt.Sprintf("%s%s%s", dockerRepoPath, "/", dir)
	if strings.Contains(dockerRepoPath, "\\") {
		hostDir = fmt.Sprintf("%s%s%s", dockerRepoPath, "\\", dir)
	}
	log.WithFields(log.Fields{
		"dockerRepoPath": dockerRepoPath,
		"dir":            dir,
		"localPath":      localPath,
		"hostDir":        hostDir,
	}).Trace("getHostPath")
	return hostDir
}

// detectDocker checks if the current process is running inside a Docker container
// and returns whether it's running in Docker and the container ID
func detectDocker() (bool, string, error) {
	// First check for .dockerenv as a quick test
	if _, err := os.Stat("/.dockerenv"); err == nil {
		// Get container ID from mountinfo
		id, err := getContainerIDFromMountInfo()
		if err != nil {
			return true, "", fmt.Errorf("container ID not found: %v", err)
		}
		return true, id, nil
	}

	// Double check mountinfo for docker mounts as backup method
	id, err := getContainerIDFromMountInfo()
	if err == nil && id != "" {
		return true, id, nil
	}

	return false, "", nil
}

// getContainerIDFromMountInfo retrieves the container ID from /proc/self/mountinfo
func getContainerIDFromMountInfo() (string, error) {
	file, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "/docker/containers/") {
			// Extract container ID from the path
			parts := strings.Split(line, "/docker/containers/")
			if len(parts) > 1 {
				// The container ID will be the first part before any subsequent slash
				containerParts := strings.Split(parts[1], "/")
				if len(containerParts) > 0 && len(containerParts[0]) >= 12 {
					return containerParts[0][:12], nil // Return first 12 chars of container ID
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading mountinfo: %v", err)
	}

	return "", errors.New("container ID not found in mountinfo")
}

// GetContainerInfo returns the current Docker context info if available
func GetContainerInfo(containerPath string) (string, error) {
	if !isDocker || containerID == "" {
		return "", nil
	}

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.WithError(err).Error("Failed to create Docker client")
		return "", err
	}
	defer cli.Close()

	containerJSON, err := cli.ContainerInspect(context.Background(), containerID)
	if err != nil {
		log.WithError(err).Error("Failed to inspect container")
		return "", err
	}

	for _, mount := range containerJSON.Mounts {
		if mount.Destination == containerPath {
			return mount.Source, nil
		}
	}

	return "", errors.New("container path not found")
}
