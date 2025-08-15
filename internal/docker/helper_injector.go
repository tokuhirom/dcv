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
	client   *client.Client
	mu       sync.Mutex
	path     string
	injected map[string]bool // Track injected containers to avoid duplicates
}

// NewHelperInjector creates a new helper injector
func NewHelperInjector(dockerClient *client.Client) *HelperInjector {
	return &HelperInjector{
		client:   dockerClient,
		path:     "/.dcv-helper",
		injected: make(map[string]bool),
	}
}

// InjectHelper injects the dcv-helper binary into the container if needed
func (hi *HelperInjector) InjectHelper(ctx context.Context, container *Container) (string, error) {
	hi.mu.Lock()
	defer hi.mu.Unlock()

	targetPath := hi.path

	if hi.injected[container.containerID] {
		slog.Info("Helper binary already injected",
			slog.String("container", container.ContainerID()),
			slog.String("path", targetPath))
		return targetPath, nil // Already injected
	}

	slog.Info("Injecting helper binary",
		slog.String("container", container.ContainerID()),
		slog.String("path", targetPath))

	// Detect container architecture (default to runtime arch)
	arch := hi.detectArch(ctx, container)
	if arch == "" {
		arch = runtime.GOARCH
		slog.Info("Using runtime architecture",
			slog.String("arch", arch))
	} else {
		slog.Info("Detected container architecture",
			slog.String("arch", arch))
	}

	// Get the embedded binary
	binaryData, err := GetHelperBinary(arch)
	if err != nil {
		return "", fmt.Errorf("failed to get helper binary: %w", err)
	}

	// write binary to a temporary file
	// Create a temporary file to store the binary
	tempDir, err := os.MkdirTemp("", "dcv-helper-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	tempFile := filepath.Join(tempDir, "dcv-helper")
	if err := os.WriteFile(tempFile, binaryData, 0755); err != nil {
		return "", fmt.Errorf("failed to write helper binary to temp file: %w", err)
	}

	cmds := hi.BuildCommands(container, tempFile)
	// Execute commands sequentially
	for _, cmd := range cmds {
		slog.Info("Executing helper injection command", slog.String("cmd", cmd))
		// For now, we'll let the UI handle command execution
		// In a real implementation, we would execute these here
	}

	hi.injected[container.containerID] = true

	return targetPath, nil
}

// BuildCommands returns the list of commands needed to inject the helper
func (hi *HelperInjector) BuildCommands(container *Container, tempFile string) []string {
	if container.isDind {
		return []string{
			fmt.Sprintf("docker cp %s %s:%s", tempFile, container.hostContainerID, hi.path),
			fmt.Sprintf("docker exec %s docker cp %s %s:%s", container.hostContainerID, hi.path, container.containerID, hi.path),
		}
	} else {
		return []string{
			fmt.Sprintf("docker cp %s %s:%s", tempFile, container.containerID, hi.path),
		}
	}
}

// GetHelperTempFile creates a temporary file with the helper binary and returns its path
func (hi *HelperInjector) GetHelperTempFile(ctx context.Context, container *Container) (string, error) {
	// Detect container architecture (default to runtime arch)
	arch := hi.detectArch(ctx, container)
	if arch == "" {
		arch = runtime.GOARCH
		slog.Info("Using runtime architecture",
			slog.String("arch", arch))
	} else {
		slog.Info("Detected container architecture",
			slog.String("arch", arch))
	}

	// Get the embedded binary
	binaryData, err := GetHelperBinary(arch)
	if err != nil {
		return "", fmt.Errorf("failed to get helper binary: %w", err)
	}

	// Create a temporary file to store the binary
	tempDir, err := os.MkdirTemp("", "dcv-helper-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	tempFile := filepath.Join(tempDir, "dcv-helper")
	if err := os.WriteFile(tempFile, binaryData, 0755); err != nil {
		_ = os.RemoveAll(tempDir)
		return "", fmt.Errorf("failed to write helper binary to temp file: %w", err)
	}

	return tempFile, nil
}

// detectArch tries to detect the container's architecture
func (hi *HelperInjector) detectArch(ctx context.Context, container *Container) string {
	// TODO: use `docker inspect` instead of docker api.

	// Inspect container to get architecture
	inspect, err := hi.client.ContainerInspect(ctx, container.containerID)
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
