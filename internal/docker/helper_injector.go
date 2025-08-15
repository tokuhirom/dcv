package docker

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/docker/docker/client"
)

// HelperInjector manages injecting helper binaries into containers
type HelperInjector struct {
	client       *client.Client
	injectedBins map[string]string // container -> binary path
	mu           sync.Mutex
}

// NewHelperInjector creates a new helper injector
func NewHelperInjector(dockerClient *client.Client) *HelperInjector {
	return &HelperInjector{
		client:       dockerClient,
		injectedBins: make(map[string]string),
	}
}

// InjectHelper injects the dcv-helper binary into the container if needed
func (hi *HelperInjector) InjectHelper(ctx context.Context, containerID string) (string, error) {
	hi.mu.Lock()
	defer hi.mu.Unlock()

	targetPath := "/tmp/.dcv-helper"

	// Check if already injected and still exists
	if binPath, exists := hi.injectedBins[containerID]; exists && binPath != "" {
		// Verify it still exists in the container
		if hi.helperExists(ctx, containerID, binPath) {
			slog.Debug("Helper already injected", "container", containerID, "path", binPath)
			return binPath, nil
		}
		// Helper was removed, need to re-inject
		delete(hi.injectedBins, containerID)
	}

	slog.Info("Injecting helper binary", "container", containerID, "path", targetPath)

	// Detect container architecture (default to runtime arch)
	arch := hi.detectArch(ctx, containerID)
	if arch == "" {
		arch = runtime.GOARCH
		slog.Debug("Using runtime architecture", "arch", arch)
	}

	// Get the embedded binary
	binaryData, err := GetHelperBinary(arch)
	if err != nil {
		return "", fmt.Errorf("failed to get helper binary: %w", err)
	}

	// Create tar archive with the binary
	if err := hi.copyToContainer(ctx, containerID, binaryData, targetPath); err != nil {
		return "", fmt.Errorf("failed to copy helper to container: %w", err)
	}

	// Make it executable
	if err := hi.makeExecutable(ctx, containerID, targetPath); err != nil {
		slog.Warn("Failed to make helper executable with chmod, trying alternative", "error", err)
		// Some containers don't have chmod, but the file might still be executable
	}

	// Record the injection
	hi.injectedBins[containerID] = targetPath

	slog.Info("Helper binary injected successfully", "container", containerID, "path", targetPath)
	return targetPath, nil
}

// helperExists checks if the helper binary exists in the container
func (hi *HelperInjector) helperExists(ctx context.Context, containerID, path string) bool {
	// Try to run the helper's version command using docker exec
	output, err := ExecuteCaptured("docker", "exec", containerID, path, "version")
	if err != nil {
		return false
	}

	// Check if we got version output
	return strings.Contains(string(output), "dcv-helper")
}

// detectArch tries to detect the container's architecture
func (hi *HelperInjector) detectArch(ctx context.Context, containerID string) string {
	// Inspect container to get architecture
	inspect, err := hi.client.ContainerInspect(ctx, containerID)
	if err != nil {
		slog.Debug("Failed to inspect container for architecture", "error", err)
		return ""
	}

	// Architecture is in format like "amd64", "arm64", etc.
	if inspect.Platform != "" {
		// Platform might be like "linux/amd64"
		parts := strings.Split(inspect.Platform, "/")
		if len(parts) > 1 {
			return parts[1]
		}
	}

	// Try to get from image config
	if inspect.Config.Labels != nil {
		if arch, ok := inspect.Config.Labels["architecture"]; ok {
			return arch
		}
	}

	// Default detection failed
	return ""
}

// copyToContainer copies binary data to a container using docker cp command
func (hi *HelperInjector) copyToContainer(ctx context.Context, containerID string, binaryData []byte, targetPath string) error {
	// Create a temporary file to store the binary
	tempDir, err := os.MkdirTemp("", "dcv-helper-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	tempFile := filepath.Join(tempDir, "dcv-helper")
	if err := os.WriteFile(tempFile, binaryData, 0755); err != nil {
		return fmt.Errorf("failed to write helper binary to temp file: %w", err)
	}

	// Use docker cp to copy the file to the container
	output, err := ExecuteCaptured("docker", "cp", tempFile, fmt.Sprintf("%s:%s", containerID, targetPath))
	if err != nil {
		return fmt.Errorf("docker cp failed: %w, output: %s", err, string(output))
	}

	slog.Debug("Successfully copied helper to container using docker cp", "container", containerID, "path", targetPath)
	return nil
}

// makeExecutable tries to make the file executable in the container
func (hi *HelperInjector) makeExecutable(ctx context.Context, containerID, path string) error {
	// Use docker exec to run chmod
	output, err := ExecuteCaptured("docker", "exec", containerID, "chmod", "+x", path)
	if err != nil {
		slog.Debug("Failed to make helper executable", "error", err, "output", string(output))
		return fmt.Errorf("chmod failed: %w", err)
	}
	return nil
}

// Cleanup removes the injected helper from a container
func (hi *HelperInjector) Cleanup(ctx context.Context, containerID string) error {
	hi.mu.Lock()
	defer hi.mu.Unlock()

	binPath, exists := hi.injectedBins[containerID]
	if !exists {
		return nil
	}

	// Try to remove the binary using docker exec
	_, _ = ExecuteCaptured("docker", "exec", containerID, "rm", "-f", binPath)

	delete(hi.injectedBins, containerID)
	slog.Debug("Cleaned up helper binary", "container", containerID, "path", binPath)
	return nil
}

// CleanupAll removes all injected helpers
func (hi *HelperInjector) CleanupAll(ctx context.Context) {
	hi.mu.Lock()
	containers := make([]string, 0, len(hi.injectedBins))
	for c := range hi.injectedBins {
		containers = append(containers, c)
	}
	hi.mu.Unlock()

	for _, containerID := range containers {
		_ = hi.Cleanup(ctx, containerID)
	}
}
