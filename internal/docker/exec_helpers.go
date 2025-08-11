package docker

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

// ExecuteInContainer executes a command in a container and returns the output
func ExecuteInContainer(ctx context.Context, dockerClient *client.Client, containerID string, cmd []string) (string, error) {
	// Create exec instance
	execConfig := container.ExecOptions{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          false,
	}

	resp, err := dockerClient.ContainerExecCreate(ctx, containerID, execConfig)
	if err != nil {
		return "", fmt.Errorf("failed to create exec: %w", err)
	}

	// Attach to the exec instance
	attachResp, err := dockerClient.ContainerExecAttach(ctx, resp.ID, container.ExecStartOptions{
		Tty: false,
	})
	if err != nil {
		return "", fmt.Errorf("failed to attach to exec: %w", err)
	}
	defer attachResp.Close()

	// Read the output
	var stdout, stderr bytes.Buffer
	if execConfig.Tty {
		// If TTY is enabled, output is not multiplexed
		_, err = io.Copy(&stdout, attachResp.Reader)
	} else {
		// Demultiplex stdout and stderr
		_, err = stdcopy.StdCopy(&stdout, &stderr, attachResp.Reader)
	}
	if err != nil {
		return "", fmt.Errorf("failed to read output: %w", err)
	}

	// Check exec exit code
	inspectResp, err := dockerClient.ContainerExecInspect(ctx, resp.ID)
	if err != nil {
		return "", fmt.Errorf("failed to inspect exec: %w", err)
	}

	if inspectResp.ExitCode != 0 {
		errOutput := stderr.String()
		if errOutput == "" {
			errOutput = stdout.String()
		}
		return "", fmt.Errorf("command failed with exit code %d: %s", inspectResp.ExitCode, errOutput)
	}

	return stdout.String(), nil
}
