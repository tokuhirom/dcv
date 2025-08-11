package docker

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"runtime"
	"strings"
	"sync"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
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
	// Try to run the helper's version command
	cmd := []string{path, "version"}
	resp, err := hi.client.ContainerExecCreate(ctx, containerID, container.ExecOptions{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
	})
	if err != nil {
		return false
	}

	attach, err := hi.client.ContainerExecAttach(ctx, resp.ID, container.ExecStartOptions{})
	if err != nil {
		return false
	}
	defer attach.Close()

	// Read output to see if it works
	var output bytes.Buffer
	_, err = stdcopy.StdCopy(&output, io.Discard, attach.Reader)
	if err != nil {
		return false
	}

	// Check if we got version output
	return strings.Contains(output.String(), "dcv-helper")
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

// copyToContainer copies binary data to a container as a tar archive
func (hi *HelperInjector) copyToContainer(ctx context.Context, containerID string, binaryData []byte, targetPath string) error {
	// Create tar archive in memory
	var tarBuffer bytes.Buffer
	tarWriter := tar.NewWriter(&tarBuffer)

	// Remove leading slash for tar entry
	tarPath := strings.TrimPrefix(targetPath, "/")

	// Add binary to tar
	header := &tar.Header{
		Name: tarPath,
		Mode: 0755,
		Size: int64(len(binaryData)),
	}

	if err := tarWriter.WriteHeader(header); err != nil {
		return fmt.Errorf("failed to write tar header: %w", err)
	}

	if _, err := tarWriter.Write(binaryData); err != nil {
		return fmt.Errorf("failed to write binary to tar: %w", err)
	}

	if err := tarWriter.Close(); err != nil {
		return fmt.Errorf("failed to close tar writer: %w", err)
	}

	// Copy tar to container
	return hi.client.CopyToContainer(ctx, containerID, "/", &tarBuffer, container.CopyToContainerOptions{})
}

// makeExecutable tries to make the file executable in the container
func (hi *HelperInjector) makeExecutable(ctx context.Context, containerID, path string) error {
	// Try chmod first
	cmd := []string{"chmod", "+x", path}
	resp, err := hi.client.ContainerExecCreate(ctx, containerID, container.ExecOptions{
		Cmd: cmd,
	})
	if err != nil {
		return err
	}

	return hi.client.ContainerExecStart(ctx, resp.ID, container.ExecStartOptions{})
}

// Cleanup removes the injected helper from a container
func (hi *HelperInjector) Cleanup(ctx context.Context, containerID string) error {
	hi.mu.Lock()
	defer hi.mu.Unlock()

	binPath, exists := hi.injectedBins[containerID]
	if !exists {
		return nil
	}

	// Try to remove the binary
	cmd := []string{"rm", "-f", binPath}
	resp, err := hi.client.ContainerExecCreate(ctx, containerID, container.ExecOptions{
		Cmd: cmd,
	})
	if err == nil {
		_ = hi.client.ContainerExecStart(ctx, resp.ID, container.ExecStartOptions{})
	}

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
