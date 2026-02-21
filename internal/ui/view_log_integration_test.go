package ui

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tokuhirom/dcv/internal/docker"
)

func skipIfNoDocker(t *testing.T) {
	t.Helper()
	if err := exec.Command("docker", "info").Run(); err != nil {
		t.Skip("Docker is not available, skipping integration test")
	}
}

func TestLogView_LongLines_Integration(t *testing.T) {
	skipIfNoDocker(t)

	containerName := "dcv-test-log-longlines"

	// Generate logs with lines of varying lengths including very long lines
	// The script outputs:
	// - Short lines (~30 chars)
	// - Medium lines (~120 chars)
	// - Very long lines (~300 chars)
	script := `
i=1
while [ $i -le 5 ]; do
  echo "short line $i"
  printf 'medium line %d: %s\n' "$i" "$(head -c 100 /dev/urandom | base64 | tr -d '\n')"
  printf 'long line %d: %s\n' "$i" "$(head -c 250 /dev/urandom | base64 | tr -d '\n')"
  i=$((i + 1))
done
`

	out, err := exec.Command("docker", "run", "--rm", "--name", containerName,
		"alpine:latest", "sh", "-c", script).CombinedOutput()
	require.NoError(t, err, "failed to run container: %s", string(out))

	logs := strings.Split(strings.TrimRight(string(out), "\n"), "\n")
	require.True(t, len(logs) >= 15, "expected at least 15 log lines, got %d", len(logs))

	// Verify we have lines of different lengths
	var hasLong bool
	for _, line := range logs {
		if len(line) > 200 {
			hasLong = true
			break
		}
	}
	require.True(t, hasLong, "expected at least one line longer than 200 chars")

	t.Run("render and scroll with real long log lines", func(t *testing.T) {
		termWidth := 80
		termHeight := 20
		container := docker.NewContainer("test-container", containerName, "exited", "exited")

		model := &Model{
			currentView: LogView,
			width:       termWidth,
			Height:      termHeight,
		}
		model.logViewModel.SwitchToLogView(model, container)
		model.logViewModel.LogLines(model, logs)

		// Render initial view
		availableHeight := termHeight - 4 // nav + title + footer
		result := model.logViewModel.render(model, availableHeight)
		assert.NotEmpty(t, result)

		// The first log line should be visible (auto-scrolled to bottom,
		// so scroll to top first)
		model.logViewModel.HandleGoToBeginning()
		result = model.logViewModel.render(model, availableHeight)
		assert.Contains(t, result, "short line 1", "first short line should be visible at top")

		// Scroll to the end and verify the last line is reachable
		model.logViewModel.HandleGoToEnd(model)
		result = model.logViewModel.render(model, availableHeight)

		lastLine := logs[len(logs)-1]
		// The last line may be very long, check for its prefix
		lastLinePrefix := lastLine[:min(50, len(lastLine))]
		assert.Contains(t, result, lastLinePrefix,
			"last log line should be visible when scrolled to end")
	})

	t.Run("line-by-line navigation covers all lines", func(t *testing.T) {
		termWidth := 80
		termHeight := 15
		container := docker.NewContainer("test-container", containerName, "exited", "exited")

		model := &Model{
			currentView: LogView,
			width:       termWidth,
			Height:      termHeight,
		}
		model.logViewModel.SwitchToLogView(model, container)
		model.logViewModel.LogLines(model, logs)

		// Start from top
		model.logViewModel.HandleGoToBeginning()

		// Collect all rendered content by scrolling line-by-line
		var allRendered strings.Builder
		availableHeight := termHeight - 4

		for range 200 { // safety limit
			result := model.logViewModel.render(model, availableHeight)
			allRendered.WriteString(result)

			prevScroll := model.logViewModel.logScrollY
			model.logViewModel.HandleDown(model)
			if model.logViewModel.logScrollY == prevScroll {
				break // Reached the end
			}
		}

		rendered := allRendered.String()

		// Every short line should have appeared somewhere
		for i := 1; i <= 5; i++ {
			assert.Contains(t, rendered, "short line "+string(rune('0'+i)),
				"short line %d should appear in scrolled output", i)
		}
	})

	t.Run("HandleDown progresses through all lines", func(t *testing.T) {
		termWidth := 80
		termHeight := 10
		container := docker.NewContainer("test-container", containerName, "exited", "exited")

		model := &Model{
			currentView: LogView,
			width:       termWidth,
			Height:      termHeight,
		}
		model.logViewModel.SwitchToLogView(model, container)
		model.logViewModel.LogLines(model, logs)
		model.logViewModel.HandleGoToBeginning()

		// Scroll down line by line until we reach the end
		maxScroll := model.logViewModel.calculateMaxScroll(model)
		for range maxScroll + 10 { // extra iterations to ensure we stop
			prevScroll := model.logViewModel.logScrollY
			model.logViewModel.HandleDown(model)
			if model.logViewModel.logScrollY == prevScroll {
				break
			}
		}

		// Should have reached max scroll
		assert.Equal(t, maxScroll, model.logViewModel.logScrollY,
			"should reach max scroll position by pressing down repeatedly")

		// Render at max scroll and verify last line is visible
		availableHeight := termHeight - 4
		result := model.logViewModel.render(model, availableHeight)
		lastLine := logs[len(logs)-1]
		lastLinePrefix := lastLine[:min(50, len(lastLine))]
		assert.Contains(t, result, lastLinePrefix,
			"last line should be visible at max scroll")
	})
}
