package docker

import (
	"errors"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
	"time"
)

func Execute(args ...string) *exec.Cmd {
	slog.Info("Executing docker command",
		slog.String("args", strings.Join(args, " ")))

	return exec.Command("docker", args...)
}

func ExecuteCaptured(args ...string) ([]byte, error) {
	cmd := Execute(args...)

	startTime := time.Now()
	cmdStr := strings.Join(cmd.Args, " ")

	output, err := cmd.CombinedOutput()
	duration := time.Since(startTime)

	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			exitCode := exitErr.ExitCode()
			return nil, fmt.Errorf("command failed with exit code %d: %w\n%s", exitCode, err, output)
		}
		return nil, fmt.Errorf("command execution failed: %w\n%s", err, output)
	}

	slog.Info("Executed command",
		slog.String("command", cmdStr),
		slog.Duration("duration", duration),
		slog.String("output", string(output)))

	return output, nil
}
