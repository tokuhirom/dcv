package docker

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"strings"
	"time"

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
	output, err := ExecuteInContainer(ctx, fo.client, container.containerID, cmd)
	if err != nil {
		return nil, fmt.Errorf("helper ls failed: %w(%v)", err, cmd)
	}

	// Parse helper output (similar format to ls -la)
	files := parseHelperLsOutput(output)
	return files, nil
}

// GetFileContent retrieves file content from a container using multiple strategies
func (fo *FileOperations) GetFileContent(ctx context.Context, containerID, filePath string) (string, error) {
	// Strategy 1: Try docker cp first (most reliable)
	content, err := fo.getFileContentWithDockerCp(ctx, containerID, filePath)
	if err == nil {
		return content, nil
	}

	slog.Debug("Docker cp failed, trying native cat", "error", err)

	// Strategy 2: Try native cat command as fallback
	content, err = fo.getFileContentNative(ctx, containerID, filePath)
	if err == nil {
		return content, nil
	}

	slog.Debug("Native cat also failed", "error", err)

	// All strategies failed
	return "", fmt.Errorf("unable to read file: docker cp and native cat both failed")
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

// getFileContentWithDockerCp gets file content using docker cp command
func (fo *FileOperations) getFileContentWithDockerCp(ctx context.Context, containerID, filePath string) (string, error) {
	// Use docker cp to extract the file content
	// docker cp CONTAINER:PATH - outputs tar to stdout
	args := []string{"docker", "cp", fmt.Sprintf("%s:%s", containerID, filePath), "-"}
	captured, err := ExecuteCaptured(args...)
	if err != nil {
		return "", fmt.Errorf("docker cp failed: %w", err)
	}

	// Extract content from tar format
	content, err := extractFileFromTar(captured, filePath)
	if err != nil {
		return "", fmt.Errorf("failed to extract from tar: %w", err)
	}

	return content, nil
}

// parseHelperLsOutput parses the output from our helper's ls command
// The format is similar to ls -la but simplified:
// dRWXRWXRWX      4096 dirname
// -rwxr-xr-x       123 filename

// TODO: helper に`ls -la` と同等の出力をさせる｡
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

// extractFileFromTar extracts file content from tar archive
func extractFileFromTar(tarData []byte, filePath string) (string, error) {
	// Create a tar reader
	tarReader := tar.NewReader(bytes.NewReader(tarData))

	// Get just the filename from the path for matching
	fileName := filepath.Base(filePath)

	// Read through the tar archive
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("error reading tar: %w", err)
		}

		// Check if this is the file we're looking for
		// The tar might have just the filename without the full path
		if header.Name == fileName || header.Name == filePath || filepath.Base(header.Name) == fileName {
			// Read the file content
			content, err := io.ReadAll(tarReader)
			if err != nil {
				return "", fmt.Errorf("error reading file from tar: %w", err)
			}
			return string(content), nil
		}
	}

	return "", fmt.Errorf("file not found in tar archive: %s", filePath)
}
