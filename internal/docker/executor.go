package docker

import (
	"errors"
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

	exitCode := 0
	errorStr := ""

	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			exitCode = exitErr.ExitCode()
		}
		errorStr = err.Error()
	}

	slog.Info("Executed command",
		slog.String("command", cmdStr),
		slog.Int("exitCode", exitCode),
		slog.String("error", errorStr),
		slog.Duration("duration", duration),
		slog.String("output", string(output)))

	return output, err
}
