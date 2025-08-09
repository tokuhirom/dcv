package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type CommandViewModel struct {
	commandMode       bool
	commandBuffer     string
	commandHistory    []string
	commandHistoryIdx int
	commandCursorPos  int
}

func (m *CommandViewModel) Start() {
	m.commandMode = true
	m.commandBuffer = ":"
	m.commandCursorPos = 1
}

func (m *CommandViewModel) HandleKeys(model *Model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		// Exit command mode
		m.commandMode = false
		m.commandBuffer = ""
		m.commandCursorPos = 0
		return model, nil

	case tea.KeyEnter:
		// Execute command
		return m.executeCommand(model)

	case tea.KeyBackspace, tea.KeyCtrlH:
		if len(m.commandBuffer) > 1 && m.commandCursorPos > 1 {
			m.commandBuffer = m.commandBuffer[:m.commandCursorPos-1] + m.commandBuffer[m.commandCursorPos:]
			m.commandCursorPos--
		}
		return model, nil

	case tea.KeyLeft, tea.KeyCtrlB:
		if m.commandCursorPos > 1 {
			m.commandCursorPos--
		}
		return model, nil

	case tea.KeyRight, tea.KeyCtrlF:
		if m.commandCursorPos < len(m.commandBuffer) {
			m.commandCursorPos++
		}
		return model, nil

	case tea.KeyUp, tea.KeyCtrlP:
		// Navigate command history
		if m.commandHistoryIdx > 0 {
			m.commandHistoryIdx--
			if m.commandHistoryIdx < len(m.commandHistory) {
				m.commandBuffer = ":" + m.commandHistory[m.commandHistoryIdx]
				m.commandCursorPos = len(m.commandBuffer)
			}
		}
		return model, nil

	case tea.KeyDown, tea.KeyCtrlN:
		// Navigate command history
		if m.commandHistoryIdx < len(m.commandHistory)-1 {
			m.commandHistoryIdx++
			m.commandBuffer = ":" + m.commandHistory[m.commandHistoryIdx]
			m.commandCursorPos = len(m.commandBuffer)
		} else if m.commandHistoryIdx == len(m.commandHistory)-1 {
			m.commandHistoryIdx++
			m.commandBuffer = ":"
			m.commandCursorPos = 1
		}
		return model, nil

	default:
		switch {
		case msg.Type == tea.KeyRunes:
			m.commandBuffer = m.commandBuffer[:m.commandCursorPos] + msg.String() + m.commandBuffer[m.commandCursorPos:]
			m.commandCursorPos += len(msg.String())
		case msg.Type == tea.KeySpace:
			m.commandBuffer = m.commandBuffer[:m.commandCursorPos] + " " + m.commandBuffer[m.commandCursorPos:]
			m.commandCursorPos++
		}
		return model, nil
	}
}

func (m *CommandViewModel) executeCommand(model *Model) (tea.Model, tea.Cmd) {
	command := strings.TrimSpace(m.commandBuffer[1:]) // Remove leading ':'

	// Add to command history
	if command != "" && (len(m.commandHistory) == 0 || m.commandHistory[len(m.commandHistory)-1] != command) {
		m.commandHistory = append(m.commandHistory, command)
	}
	m.commandHistoryIdx = len(m.commandHistory)

	// Exit command mode
	m.commandMode = false
	m.commandBuffer = ""
	m.commandCursorPos = 0

	// Parse and execute command
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return model, nil
	}

	switch parts[0] {
	case "q", "quit":
		// Show quit confirmation
		model.quitConfirmation = true
		return model, nil

	case "q!", "quit!":
		// Force quit without confirmation
		return model, tea.Quit

	case "h", "help":
		return model, model.helpViewModel.Show(model, model.currentView)

	default:
		// Try to execute as a key handler command
		return m.executeKeyHandlerCommand(model, parts[0])
	}
}

// executeKeyHandlerCommand executes a command by name
func (m *CommandViewModel) executeKeyHandlerCommand(model *Model, cmdName string) (tea.Model, tea.Cmd) {
	cmd, exists := model.viewCommandRegistry[model.currentView][cmdName]
	if !exists {
		if _, ok := model.allCommands[cmdName]; ok {
			model.err = fmt.Errorf(":%s is not available in %s", cmdName, model.currentView.String())
		} else {
			model.err = fmt.Errorf("unknown command: :%s", cmdName)
		}
		return model, nil
	}

	// Execute the command
	return cmd.Handler(tea.KeyMsg{})
}

func (m *CommandViewModel) RenderCmdLine() string {
	cursor := " "
	if m.commandCursorPos < len(m.commandBuffer) {
		cursor = string(m.commandBuffer[m.commandCursorPos])
	}

	// Build command line with cursor
	before := m.commandBuffer[:m.commandCursorPos]
	after := ""
	if m.commandCursorPos < len(m.commandBuffer) {
		after = m.commandBuffer[m.commandCursorPos+1:]
	}

	cursorStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("226")).
		Foreground(lipgloss.Color("235"))

	return before + cursorStyle.Render(cursor) + after
}
