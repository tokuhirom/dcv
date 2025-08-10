package terminal

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// Terminal represents a terminal emulator
type Terminal struct {
	Name    string
	Command string
	Args    func(cmd string) []string
}

// DetectTerminal detects available terminal emulator
func DetectTerminal() (*Terminal, error) {
	switch runtime.GOOS {
	case "darwin":
		return detectMacTerminal()
	case "linux":
		return detectLinuxTerminal()
	case "windows":
		return detectWindowsTerminal()
	default:
		return nil, fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

// detectMacTerminal detects terminal on macOS
func detectMacTerminal() (*Terminal, error) {
	terminals := []Terminal{
		{
			Name:    "iTerm2",
			Command: "osascript",
			Args: func(cmd string) []string {
				script := fmt.Sprintf(`
					tell application "iTerm"
						create window with default profile
						tell current session of current window
							write text "%s"
						end tell
					end tell
				`, escapeAppleScript(cmd))
				return []string{"-e", script}
			},
		},
		{
			Name:    "Terminal",
			Command: "osascript",
			Args: func(cmd string) []string {
				script := fmt.Sprintf(`
					tell application "Terminal"
						do script "%s"
						activate
					end tell
				`, escapeAppleScript(cmd))
				return []string{"-e", script}
			},
		},
	}

	// Check for iTerm2 first as it's preferred
	if isAppInstalled("iTerm") {
		return &terminals[0], nil
	}

	// Fall back to Terminal.app which is always available on macOS
	return &terminals[1], nil
}

// detectLinuxTerminal detects terminal on Linux
func detectLinuxTerminal() (*Terminal, error) {
	// List of terminals to try in order of preference
	terminals := []Terminal{
		{
			Name:    "gnome-terminal",
			Command: "gnome-terminal",
			Args: func(cmd string) []string {
				return []string{"--", "sh", "-c", cmd}
			},
		},
		{
			Name:    "konsole",
			Command: "konsole",
			Args: func(cmd string) []string {
				return []string{"-e", "sh", "-c", cmd}
			},
		},
		{
			Name:    "xfce4-terminal",
			Command: "xfce4-terminal",
			Args: func(cmd string) []string {
				return []string{"-e", cmd}
			},
		},
		{
			Name:    "xterm",
			Command: "xterm",
			Args: func(cmd string) []string {
				return []string{"-e", cmd}
			},
		},
		{
			Name:    "alacritty",
			Command: "alacritty",
			Args: func(cmd string) []string {
				return []string{"-e", "sh", "-c", cmd}
			},
		},
		{
			Name:    "kitty",
			Command: "kitty",
			Args: func(cmd string) []string {
				return []string{"sh", "-c", cmd}
			},
		},
		{
			Name:    "terminator",
			Command: "terminator",
			Args: func(cmd string) []string {
				return []string{"-e", cmd}
			},
		},
	}

	// Try to find an available terminal
	for _, term := range terminals {
		if _, err := exec.LookPath(term.Command); err == nil {
			return &term, nil
		}
	}

	// Try x-terminal-emulator as last resort (Debian/Ubuntu)
	if _, err := exec.LookPath("x-terminal-emulator"); err == nil {
		return &Terminal{
			Name:    "x-terminal-emulator",
			Command: "x-terminal-emulator",
			Args: func(cmd string) []string {
				return []string{"-e", cmd}
			},
		}, nil
	}

	return nil, fmt.Errorf("no supported terminal emulator found")
}

// detectWindowsTerminal detects terminal on Windows
func detectWindowsTerminal() (*Terminal, error) {
	terminals := []Terminal{
		{
			Name:    "Windows Terminal",
			Command: "wt",
			Args: func(cmd string) []string {
				return []string{"new-tab", "--", "cmd", "/c", cmd}
			},
		},
		{
			Name:    "Command Prompt",
			Command: "cmd",
			Args: func(cmd string) []string {
				return []string{"/c", "start", "cmd", "/k", cmd}
			},
		},
	}

	// Try Windows Terminal first
	if _, err := exec.LookPath("wt"); err == nil {
		return &terminals[0], nil
	}

	// Fall back to cmd which is always available
	return &terminals[1], nil
}

// OpenInNewWindow opens a command in a new terminal window
func OpenInNewWindow(containerID string, command []string) error {
	terminal, err := DetectTerminal()
	if err != nil {
		return err
	}

	// Build the docker exec command
	dockerCmd := buildDockerCommand(containerID, command)

	// Get the arguments for the terminal
	args := terminal.Args(dockerCmd)

	// Execute the terminal command
	cmd := exec.Command(terminal.Command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Start()
}

// buildDockerCommand builds the docker exec command string
func buildDockerCommand(containerID string, command []string) string {
	// Build the full docker exec command
	parts := []string{"docker", "exec", "-it", containerID}
	parts = append(parts, command...)

	// Join with proper escaping
	var escapedParts []string
	for _, part := range parts {
		if strings.Contains(part, " ") {
			escapedParts = append(escapedParts, fmt.Sprintf("'%s'", part))
		} else {
			escapedParts = append(escapedParts, part)
		}
	}

	return strings.Join(escapedParts, " ")
}

// escapeAppleScript escapes a string for use in AppleScript
func escapeAppleScript(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	return s
}

// isAppInstalled checks if an application is installed on macOS
func isAppInstalled(appName string) bool {
	cmd := exec.Command("osascript", "-e", fmt.Sprintf(`tell application "System Events" to return name of every application process contains "%s"`, appName))
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(output)) == "true"
}
