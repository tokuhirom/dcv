package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/tokuhirom/dcv/internal/docker"
)

// CommandExecutionView represents the command execution output view
type CommandExecutionView struct {
	// View state
	width      int
	height     int
	output     []string
	scrollY    int
	commandStr string
	isDone     bool

	// Dependencies
	dockerClient *docker.Client
	rootScreen   tea.Model
	previousView tea.Model
}

// NewCommandExecutionView creates a new command execution view
func NewCommandExecutionView(dockerClient *docker.Client, commandStr string, previousView tea.Model) *CommandExecutionView {
	return &CommandExecutionView{
		dockerClient: dockerClient,
		commandStr:   commandStr,
		previousView: previousView,
		output:       []string{},
		isDone:       false,
	}
}

// SetRootScreen sets the root screen reference
func (v *CommandExecutionView) SetRootScreen(root tea.Model) {
	v.rootScreen = root
}

// Init initializes the view
func (v *CommandExecutionView) Init() tea.Cmd {
	// The actual command execution should be initiated by the caller
	return nil
}

// Update handles messages for this view
func (v *CommandExecutionView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height
		return v, nil

	case tea.KeyMsg:
		return v.handleKeyPress(msg)

	case commandOutputLineMsg:
		v.output = append(v.output, msg.line)
		// Auto-scroll to bottom
		maxScroll := len(v.output) - (v.height - 5)
		if maxScroll > 0 {
			v.scrollY = maxScroll
		}
		return v, nil

	case commandCompleteMsg:
		v.isDone = true
		if msg.err != nil {
			v.output = append(v.output, fmt.Sprintf("\nCommand failed: %v", msg.err))
		} else {
			v.output = append(v.output, "\nCommand completed successfully.")
		}
		// Auto-scroll to bottom
		maxScroll := len(v.output) - (v.height - 5)
		if maxScroll > 0 {
			v.scrollY = maxScroll
		}
		return v, nil
	}

	return v, nil
}

// View renders the command execution output
func (v *CommandExecutionView) View() string {
	if v.width == 0 || v.height == 0 {
		return "Loading..."
	}

	var s strings.Builder

	// Header
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("7")).
		Background(lipgloss.Color("4")).
		Width(v.width).
		Padding(0, 1).
		Render("Command Execution")
	s.WriteString(header + "\n")

	// Command string
	cmdStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("3")).
		Bold(true)
	s.WriteString(cmdStyle.Render("$ "+v.commandStr) + "\n\n")

	// Calculate visible lines
	visibleHeight := v.height - 5 // Header + command + footer
	start := v.scrollY
	end := start + visibleHeight
	if end > len(v.output) {
		end = len(v.output)
	}

	// Render output
	if len(v.output) == 0 {
		s.WriteString("Executing command...\n")
	} else {
		for i := start; i < end; i++ {
			if i < len(v.output) {
				s.WriteString(v.output[i] + "\n")
			}
		}
	}

	// Pad to fill screen
	lines := strings.Split(s.String(), "\n")
	for len(lines) < v.height-1 {
		lines = append(lines, "")
	}

	// Footer
	footerText := "↑/↓: Scroll"
	if v.isDone {
		footerText += " • Enter/ESC: Return"
	}
	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Width(v.width).
		Align(lipgloss.Center).
		Render(footerText)

	return strings.Join(lines[:v.height-1], "\n") + "\n" + footer
}

func (v *CommandExecutionView) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if v.scrollY > 0 {
			v.scrollY--
		}
		return v, nil

	case "down", "j":
		maxScroll := len(v.output) - (v.height - 5)
		if v.scrollY < maxScroll && maxScroll > 0 {
			v.scrollY++
		}
		return v, nil

	case "enter", "esc", "q":
		if v.isDone && v.rootScreen != nil && v.previousView != nil {
			// Go back to previous view
			if switcher, ok := v.rootScreen.(interface {
				SwitchScreen(tea.Model) (tea.Model, tea.Cmd)
			}); ok {
				return switcher.SwitchScreen(v.previousView)
			}
		}
		return v, nil
	}

	return v, nil
}

// Messages
type commandOutputLineMsg struct {
	line string
}

type commandCompleteMsg struct {
	err error
}

// AddOutputLine adds a line to the command output
func (v *CommandExecutionView) AddOutputLine(line string) tea.Cmd {
	return func() tea.Msg {
		return commandOutputLineMsg{line: line}
	}
}

// MarkComplete marks the command as complete
func (v *CommandExecutionView) MarkComplete(err error) tea.Cmd {
	return func() tea.Msg {
		return commandCompleteMsg{err: err}
	}
}
