package judge

import (
	"bufio"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
)

var (
	isDocker       bool
	dockerRepoPath string
)

// init runs when the package is initialized
func init() {
	var err error
	isDocker, err = checkIfDocker()
	if err != nil {
		log.WithError(err).Warn("Failed to check if running in Docker")
		isDocker = false
	}

	if isDocker {
		dockerRepoPath, err = findDockerVolumePath("/repos")
		if err != nil {
			log.WithError(err).Fatal("Failed to find Docker volume path for /repos")
			panic("Docker volume /repos must be mounted")
		}
		log.WithFields(log.Fields{
			"isDocker":       isDocker,
			"dockerRepoPath": dockerRepoPath,
		}).Info("Docker environment detected")
	}
}

// getTempDir returns the appropriate temporary directory path based on the environment
func getTempDir(prefix string) (string, error) {
	if isDocker {
		// When in Docker, create temp dir under the host path
		return os.MkdirTemp(dockerRepoPath, prefix)
	}
	// Default behavior when not in Docker
	return os.MkdirTemp("", prefix)
}

// checkIfDocker checks if the current process is running inside a Docker container
func checkIfDocker() (bool, error) {
	// Method 1: Check for .dockerenv file
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true, nil
	}

	// Method 2: Check cgroup
	file, err := os.Open("/proc/1/cgroup")
	if err != nil {
		return false, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), "docker") {
			return true, nil
		}
	}

	return false, nil
}

// findDockerVolumePath finds the host path of a mounted Docker volume
func findDockerVolumePath(containerPath string) (string, error) {
	// Read mountinfo to find the host path
	file, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 5 {
			continue
		}

		mountPoint := fields[4]
		if mountPoint == containerPath {
			// The source path is typically in the 3rd field
			sourcePath := fields[3]
			if sourcePath == "" {
				return "", errors.New("empty source path for volume")
			}
			return sourcePath, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading mountinfo: %v", err)
	}

	return "", fmt.Errorf("volume path %s not found in mountinfo", containerPath)
}
