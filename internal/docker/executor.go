package docker

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"strings"
	"time"

	"github.com/mattn/go-runewidth"
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

	slog.Debug("Executed command",
		slog.String("command", cmdStr),
		slog.Duration("duration", duration),
		slog.String("output", runewidth.Truncate(string(output), 144, "...")))

	return output, nil
}

// ExecuteStreamingCommand executes a docker command and returns a reader for streaming output
func ExecuteStreamingCommand(ctx context.Context, args ...string) (io.ReadCloser, error) {
	slog.Info("Executing docker streaming command",
		slog.String("args", strings.Join(args, " ")))

	cmd := exec.CommandContext(ctx, "docker", args...)

	// Get stdout pipe for streaming
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	// Get stderr pipe and merge with stdout
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start command: %w", err)
	}

	// Create a multi-reader that combines stdout and stderr
	reader := io.MultiReader(stdout, stderr)

	// Return a custom ReadCloser that waits for the command to finish
	return &streamingReader{
		reader: reader,
		cmd:    cmd,
	}, nil
}

// streamingReader wraps a reader and command for proper cleanup
type streamingReader struct {
	reader io.Reader
	cmd    *exec.Cmd
}

func (sr *streamingReader) Read(p []byte) (n int, err error) {
	return sr.reader.Read(p)
}

func (sr *streamingReader) Close() error {
	// Wait for the command to finish
	return sr.cmd.Wait()
}
