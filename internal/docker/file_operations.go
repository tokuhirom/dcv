package docker

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/docker/docker/client"

	"github.com/tokuhirom/dcv/internal/models"
)

// FileOperations provides multi-strategy file operations for containers
type FileOperations struct {
	client   *client.Client
	injector *HelperInjector
}

// NewFileOperations creates a new file operations handler
func NewFileOperations(dockerClient *client.Client) *FileOperations {
	return &FileOperations{
		client:   dockerClient,
		injector: NewHelperInjector(dockerClient),
	}
}

// ListFiles lists files in a container path using multiple strategies
func (fo *FileOperations) ListFiles(ctx context.Context, containerID, path string) ([]models.ContainerFile, error) {
	// Strategy 1: Try native ls command first
	files, err := fo.listFilesNative(ctx, containerID, path)
	if err == nil {
		return files, nil
	}

	slog.Debug("Native ls failed, trying helper injection", "error", err)

	// Strategy 2: Try helper injection
	files, err = fo.listFilesWithHelper(ctx, containerID, path)
	if err == nil {
		return files, nil
	}

	slog.Debug("Helper injection failed", "error", err)

	// All strategies failed
	return nil, fmt.Errorf("unable to list files: native ls and helper injection both failed")
}

// listFilesNative tries to list files using the native ls command
func (fo *FileOperations) listFilesNative(ctx context.Context, containerID, path string) ([]models.ContainerFile, error) {
	// Execute ls -la command
	cmd := []string{"ls", "-la", path}
	output, err := ExecuteInContainer(ctx, fo.client, containerID, cmd)
	if err != nil {
		return nil, fmt.Errorf("native ls failed: %w", err)
	}

	// Parse the output
	files := models.ParseLsOutput(output)
	if len(files) == 0 && output != "" {
		return nil, fmt.Errorf("failed to parse ls output")
	}

	return files, nil
}

// listFilesWithHelper lists files using the injected helper binary
func (fo *FileOperations) listFilesWithHelper(ctx context.Context, containerID, path string) ([]models.ContainerFile, error) {
	// Inject helper if needed
	helperPath, err := fo.injector.InjectHelper(ctx, containerID)
	if err != nil {
		return nil, fmt.Errorf("failed to inject helper: %w", err)
	}

	// Execute helper ls command
	cmd := []string{helperPath, "ls", path}
	output, err := ExecuteInContainer(ctx, fo.client, containerID, cmd)
	if err != nil {
		return nil, fmt.Errorf("helper ls failed: %w", err)
	}

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
	cmd := []string{"cat", filePath}
	output, err := ExecuteInContainer(ctx, fo.client, containerID, cmd)
	if err != nil {
		return "", fmt.Errorf("native cat failed: %w", err)
	}
	return output, nil
}

// getFileContentWithHelper gets file content using the injected helper binary
func (fo *FileOperations) getFileContentWithHelper(ctx context.Context, containerID, filePath string) (string, error) {
	// Inject helper if needed
	helperPath, err := fo.injector.InjectHelper(ctx, containerID)
	if err != nil {
		return "", fmt.Errorf("failed to inject helper: %w", err)
	}

	// Execute helper cat command
	cmd := []string{helperPath, "cat", filePath}
	output, err := ExecuteInContainer(ctx, fo.client, containerID, cmd)
	if err != nil {
		return "", fmt.Errorf("helper cat failed: %w", err)
	}

	return output, nil
}

// parseHelperLsOutput parses the output from our helper's ls command
// The format is similar to ls -la but simplified:
// dRWXRWXRWX      4096 dirname
// -rwxr-xr-x       123 filename
func parseHelperLsOutput(output string) []models.ContainerFile {
	var files []models.ContainerFile
	lines := strings.Split(strings.TrimSpace(output), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		// Parse helper output format
		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}

		file := models.ContainerFile{
			Permissions: parts[0][1:], // Skip the type character
			Mode:        parts[0],
			IsDir:       parts[0][0] == 'd',
		}

		// Parse size
		var size int64
		_, _ = fmt.Sscanf(parts[1], "%d", &size)
		file.Size = size

		// Get filename (everything from parts[2] onwards)
		file.Name = strings.Join(parts[2:], " ")

		// Handle symlinks (file -> target)
		if strings.Contains(file.Name, " -> ") {
			linkParts := strings.Split(file.Name, " -> ")
			file.Name = linkParts[0]
			if len(linkParts) > 1 {
				file.LinkTarget = linkParts[1]
			}
		}

		// Set modification time to current time (helper doesn't provide timestamps)
		file.ModTime = time.Now()

		files = append(files, file)
	}

	return files
}

// Cleanup removes any injected helpers from the specified container
func (fo *FileOperations) Cleanup(ctx context.Context, containerID string) error {
	return fo.injector.Cleanup(ctx, containerID)
}

// CleanupAll removes all injected helpers from all containers
func (fo *FileOperations) CleanupAll(ctx context.Context) {
	fo.injector.CleanupAll(ctx)
}
