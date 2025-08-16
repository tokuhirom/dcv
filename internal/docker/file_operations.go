package docker

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/docker/docker/client"

	"github.com/tokuhirom/dcv/internal/models"
)

// FileOperations provides multi-strategy file operations for containers
type FileOperations struct {
	client *client.Client
}

// NewFileOperations creates a new file operations handler
func NewFileOperations(dockerClient *client.Client) *FileOperations {
	return &FileOperations{
		client: dockerClient,
	}
}

// ListFiles lists files in a container path using multiple strategies
func (fo *FileOperations) ListFiles(ctx context.Context, container *Container, path string) ([]models.ContainerFile, error) {
	// Strategy 1: Try native ls command first
	files, errNative := fo.listFilesNative(container, path)
	if errNative == nil {
		return files, nil
	}

	slog.Debug("Native ls failed, trying helper injection", "error", errNative)

	// Strategy 2: Try helper injection
	files, errHelper := fo.listFilesWithHelper(ctx, container, path)
	if errHelper == nil {
		return files, nil
	}

	slog.Debug("Helper injection failed", slog.Any("error", errHelper))

	// All strategies failed
	return nil, fmt.Errorf("unable to list files: native ls and helper injection both failed\nContainer: %s, Path: %s\nnative: %s\nhelper: %s",
		container.containerID, path, errNative, errHelper)
}

// listFilesNative tries to list files using the native ls command
func (fo *FileOperations) listFilesNative(container *Container, path string) ([]models.ContainerFile, error) {
	// Execute ls -la command
	args := container.OperationArgs("exec", "ls", "-la", path)
	captured, err := ExecuteCaptured(args...)
	if err != nil {
		return nil, fmt.Errorf("native ls failed: %w", err)
	}
	output := string(captured)

	// Parse the output
	files := models.ParseLsOutput(output)
	if len(files) == 0 && output != "" {
		return nil, fmt.Errorf("failed to parse ls output")
	}

	return files, nil
}

// listFilesWithHelper lists files using the injected helper binary
func (fo *FileOperations) listFilesWithHelper(ctx context.Context, container *Container, path string) ([]models.ContainerFile, error) {
	// Use the helper path directly
	helperPath := GetHelperPath()

	// Execute helper ls command
	cmd := []string{helperPath, "ls", path}
	args := container.OperationArgs("exec", cmd...)
	outputBytes, err := ExecuteCaptured(args...)
	if err != nil {
		return nil, fmt.Errorf("helper ls failed: %w(%v)", err, cmd)
	}
	output := string(outputBytes)

	// Parse helper output (similar format to ls -la)
	files := parseHelperLsOutput(output)
	return files, nil
}

// GetFileContent retrieves file content from a container using multiple strategies
func (fo *FileOperations) GetFileContent(ctx context.Context, containerID, filePath string) (string, error) {
	// Strategy 1: Try native cat command first
	content, err := fo.getFileContentNative(ctx, containerID, filePath)
	if err == nil {
		return content, nil
	}

	slog.Debug("Native cat failed, trying helper injection", "error", err)

	// Strategy 2: Try helper injection
	content, err = fo.getFileContentWithHelper(ctx, containerID, filePath)
	if err == nil {
		return content, nil
	}

	slog.Debug("Helper injection failed", "error", err)

	// All strategies failed
	return "", fmt.Errorf("unable to read file: native cat and helper injection both failed")
}

// getFileContentNative tries to get file content using the native cat command
func (fo *FileOperations) getFileContentNative(ctx context.Context, containerID, filePath string) (string, error) {
	// Create a temporary container for executing the cat command
	container := NewContainer(containerID, "", "", "")
	args := container.OperationArgs("exec", "cat", filePath)
	outputBytes, err := ExecuteCaptured(args...)
	if err != nil {
		return "", fmt.Errorf("native cat failed: %w", err)
	}
	return string(outputBytes), nil
}

// getFileContentWithHelper gets file content using docker cp command
func (fo *FileOperations) getFileContentWithHelper(ctx context.Context, containerID, filePath string) (string, error) {
	// Use docker cp to extract the file content
	// docker cp CONTAINER:PATH - outputs to stdout
	args := []string{"docker", "cp", fmt.Sprintf("%s:%s", containerID, filePath), "-"}
	captured, err := ExecuteCaptured(args...)
	if err != nil {
		return "", fmt.Errorf("docker cp failed: %w", err)
	}

	// The output is in tar format, so we need to extract the actual file content
	// For simplicity, we'll just return the raw output for now
	// TODO: Properly extract content from tar format
	return string(captured), nil
}

// parseHelperLsOutput parses the output from our helper's ls command
// The format is now similar to ls -la:
// drwxr-xr-x  2 uid gid 4096 Jan 1 00:00 dirname
func parseHelperLsOutput(output string) []models.ContainerFile {
	// The helper now outputs in ls -la format, so we can reuse the standard parser
	return models.ParseLsOutput(output)
}
